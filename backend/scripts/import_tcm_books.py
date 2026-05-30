"""
Batch import TCM ancient books (txt) into Qdrant.

Features:
  - Walks a directory recursively, processes all .txt files
  - Extracts book title from filename (e.g. "013-本草纲目.txt" → "本草纲目")
  - Semantic chunking: splits on TCM structure markers, then paragraphs, then sentences
  - Resume support: skips files already indexed (tracked via checkpoint file)
  - Progress bar with per-file and overall stats

Usage:
  python import_tcm_books.py --dir "D:/古籍/TCM-Ancient-Books"
  python import_tcm_books.py --dir "D:/古籍/TCM-Ancient-Books" --dry-run
  python import_tcm_books.py --dir "D:/古籍/TCM-Ancient-Books" --limit 10
"""

import argparse
import json
import os
import re
import sys
import time
import uuid
from pathlib import Path

import requests
from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct

# TCM text structure markers (angle-bracket tags used in these books)
SECTION_MARKERS = re.compile(
    r'<(?:篇名|目录|卷名|章节|标题|正文|附录|序|跋|凡例|引用|注释|按语|方名|药名|'
    r'症状|脉象|治法|禁忌|炮制|产地|采集|性味|归经|功效|主治|用法|用量|附方|'
    r'书名|作者|年代|版本|内容|原文|译文|校注)>'
)


def parse_args():
    p = argparse.ArgumentParser(description="Batch import TCM ancient books into Qdrant")
    p.add_argument("--dir", required=True, help="Directory containing TCM .txt books")
    p.add_argument("--chunk-size", type=int, default=600, help="Target chunk size in chars (default: 600)")
    p.add_argument("--chunk-overlap", type=int, default=80, help="Overlap between chunks (default: 80)")
    p.add_argument("--embed-url", default="http://127.0.0.1:8081/embed")
    p.add_argument("--qdrant-host", default="localhost")
    p.add_argument("--qdrant-port", type=int, default=6334)
    p.add_argument("--collection", default="tcm_knowledge")
    p.add_argument("--batch-size", type=int, default=30)
    p.add_argument("--limit", type=int, default=0, help="Only process first N books (0=all)")
    p.add_argument("--checkpoint-file", default="import_checkpoint.json")
    p.add_argument("--dry-run", action="store_true", help="Show files and chunks without indexing")
    return p.parse_args()


def extract_title(filepath: str) -> str:
    """Extract book title from filename: '013-本草纲目.txt' → '本草纲目'"""
    name = Path(filepath).stem
    # Remove leading number prefix like "013-" or "013－"
    name = re.sub(r'^[\d]+[－\-_]*', '', name)
    return name.strip()


def read_txt(filepath: str) -> str:
    """Read txt file, trying UTF-8 first, then common encodings."""
    encodings = ['utf-8', 'gb18030', 'gbk', 'gb2312', 'big5']
    for enc in encodings:
        try:
            with open(filepath, 'r', encoding=enc) as f:
                return f.read()
        except (UnicodeDecodeError, UnicodeError):
            continue
    # Fallback: errors='replace'
    with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
        return f.read()


def chunk_text(text: str, chunk_size: int, overlap: int) -> list[str]:
    """
    Semantic chunking for TCM texts:
      1. Split on section markers (e.g. <篇名>, <目录>)
      2. Within each section, split on double newlines (paragraphs)
      3. Merge short paragraphs, split oversized ones at sentence boundaries
    """
    # Step 0: split on TCM structure markers
    sections = SECTION_MARKERS.split(text)

    all_chunks: list[str] = []
    for section in sections:
        section = section.strip()
        if not section:
            continue

        # Step 1: split into paragraphs
        paragraphs = [p.strip() for p in section.split("\n\n") if p.strip()]

        # Step 2: merge short paragraphs
        merged: list[str] = []
        buf = ""
        for para in paragraphs:
            if len(buf) + len(para) < chunk_size:
                buf = (buf + "\n\n" + para).strip()
            else:
                if buf:
                    merged.append(buf)
                buf = para
        if buf:
            merged.append(buf)

        # Step 3: split oversized chunks at sentence boundaries
        for block in merged:
            if len(block) <= chunk_size * 1.5:
                all_chunks.append(block)
            else:
                sentences = re.split(r'(?<=[。！？；\n])\s*', block)
                buf = ""
                for s in sentences:
                    if len(buf) + len(s) < chunk_size:
                        buf += s
                    else:
                        if buf.strip():
                            all_chunks.append(buf.strip())
                        if overlap > 0 and len(buf) > overlap:
                            buf = buf[-overlap:] + s
                        else:
                            buf = s
                if buf.strip():
                    all_chunks.append(buf.strip())

    return [c for c in all_chunks if len(c) >= 30]


def embed_batch(texts: list[str], url: str, retries: int = 3) -> list[list[float]]:
    """Embed a batch of texts, with retry on failure."""
    for attempt in range(retries):
        try:
            resp = requests.post(url, json={"texts": texts}, timeout=60)
            resp.raise_for_status()
            return resp.json()["embeddings"]
        except Exception as e:
            if attempt == retries - 1:
                raise
            print(f"  Embed retry {attempt + 1}/{retries}: {e}")
            time.sleep(2)
    return []  # unreachable


def load_checkpoint(path: str) -> set[str]:
    """Load set of already-imported file paths."""
    if os.path.exists(path):
        with open(path, 'r', encoding='utf-8') as f:
            return set(json.load(f))
    return set()


def save_checkpoint(path: str, done: set[str]):
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(sorted(done), f, ensure_ascii=False, indent=2)


def collect_files(root_dir: str) -> list[str]:
    """Collect all .txt files, sorted by filename."""
    files = []
    for dirpath, _, filenames in os.walk(root_dir):
        for fn in filenames:
            if fn.lower().endswith('.txt'):
                files.append(os.path.join(dirpath, fn))
    files.sort(key=lambda p: Path(p).name)
    return files


def main():
    args = parse_args()

    files = collect_files(args.dir)
    if not files:
        print(f"No .txt files found in: {args.dir}")
        sys.exit(1)

    print(f"Found {len(files)} .txt files")
    if args.limit > 0:
        files = files[:args.limit]
        print(f"Limited to first {args.limit} files")

    if args.dry_run:
        print("\n=== DRY RUN (no indexing) ===\n")
        for i, fp in enumerate(files[:5]):
            title = extract_title(fp)
            text = read_txt(fp)
            chunks = chunk_text(text, args.chunk_size, args.chunk_overlap)
            print(f"  [{i+1}] {title}: {len(chunks)} chunks")
            for j, c in enumerate(chunks[:3]):
                print(f"      chunk {j+1}: {c[:100]}...")
        print(f"  ... ({len(files)} files total)")
        return

    # Resume support
    checkpoint_path = os.path.join(args.dir, args.checkpoint_file)
    done = load_checkpoint(checkpoint_path)
    if done:
        print(f"Resuming: {len(done)} files already indexed, {len(files) - len(done)} remaining")

    # Init Qdrant client
    client = QdrantClient(
        host=args.qdrant_host,
        port=args.qdrant_port,
        grpc_port=args.qdrant_port,
        prefer_grpc=True,
    )

    total_chunks = 0
    total_indexed = 0
    skipped = 0
    t_start = time.time()

    for i, filepath in enumerate(files):
        if filepath in done:
            skipped += 1
            continue

        title = extract_title(filepath)
        fname = Path(filepath).name
        print(f"\n[{i+1}/{len(files)}] {fname}")

        try:
            raw_text = read_txt(filepath)
            if not raw_text.strip():
                print(f"  SKIP: empty file")
                done.add(filepath)
                continue

            chunks = chunk_text(raw_text, args.chunk_size, args.chunk_overlap)
            if not chunks:
                print(f"  SKIP: no valid chunks")
                done.add(filepath)
                continue

            print(f"  {len(chunks)} chunks ({len(raw_text)} chars), embedding...")

            file_indexed = 0
            for j in range(0, len(chunks), args.batch_size):
                batch = chunks[j:j + args.batch_size]
                embeddings = embed_batch(batch, args.embed_url)

                points = []
                for text, vec in zip(batch, embeddings):
                    points.append(PointStruct(
                        id=str(uuid.uuid4()),
                        vector=vec,
                        payload={
                            "content": text,
                            "book_title": title,
                            "chapter": "",
                            "page": 0,
                        },
                    ))

                client.upsert(collection_name=args.collection, points=points)
                file_indexed += len(points)
                total_indexed += len(points)
                print(f"    {file_indexed}/{len(chunks)}", end='\r')

            print(f"    {file_indexed}/{len(chunks)} done")
            total_chunks += len(chunks)
            done.add(filepath)

            # Save checkpoint after each file
            if (i + 1) % 5 == 0:
                save_checkpoint(checkpoint_path, done)

        except Exception as e:
            print(f"  ERROR: {e}")
            save_checkpoint(checkpoint_path, done)
            print(f"  Checkpoint saved. Re-run to resume.")
            sys.exit(1)

    # Final stats
    save_checkpoint(checkpoint_path, done)
    elapsed = time.time() - t_start
    processed = len(files) - skipped

    print(f"\n{'='*50}")
    print(f"Import complete!")
    print(f"  Files processed: {processed}")
    print(f"  Files skipped:   {skipped}")
    print(f"  Total chunks:    {total_chunks}")
    print(f"  Total indexed:   {total_indexed}")
    print(f"  Elapsed:         {elapsed/60:.1f} min")

    count = client.count(collection_name=args.collection)
    print(f"  Collection '{args.collection}' total: {count.count}")
    print(f"{'='*50}")


if __name__ == "__main__":
    main()
