"""
Import a single document (PDF/TXT) into Qdrant with semantic chunking.

Chunking strategy:
  1. Split by paragraphs (double newlines)
  2. Merge short paragraphs until reaching target size
  3. Split oversized paragraphs at sentence boundaries

Usage:
  python import_document.py --file document.pdf --book-title "作物栽培学"
  python import_document.py --file document.pdf --book-title "齐民要术" --chapter "卷一"
  python import_document.py --file notes.txt --book-title "农业笔记"
  python import_document.py --file document.pdf --book-title "农书" --chunk-size 800
"""

import argparse
import re
import sys
import uuid
from pathlib import Path

import requests
from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct


def parse_args():
    p = argparse.ArgumentParser(description="Import a document into Qdrant")
    p.add_argument("--file", required=True, help="Path to document (.pdf/.txt)")
    p.add_argument("--book-title", required=True, help="Book title / source name")
    p.add_argument("--chapter", default="", help="Chapter or section name")
    p.add_argument("--chunk-size", type=int, default=600, help="Target chunk size in characters (default: 600)")
    p.add_argument("--chunk-overlap", type=int, default=100, help="Overlap between chunks in characters (default: 100)")
    p.add_argument("--start-page", type=int, default=1, help="First page to import (PDF only)")
    p.add_argument("--end-page", type=int, default=0, help="Last page to import, 0 = all (PDF only)")
    p.add_argument("--embed-url", default="http://127.0.0.1:8081/embed")
    p.add_argument("--qdrant-host", default="localhost")
    p.add_argument("--qdrant-port", type=int, default=6334)
    p.add_argument("--collection", default="agri_knowledge")
    p.add_argument("--batch-size", type=int, default=20)
    p.add_argument("--dry-run", action="store_true", help="Show chunks without indexing")
    return p.parse_args()


def extract_pdf(filepath: str, start_page: int, end_page: int) -> str:
    import pdfplumber

    full_text: list[str] = []
    with pdfplumber.open(filepath) as pdf:
        total = len(pdf.pages)
        stop = end_page if end_page > 0 else total
        stop = min(stop, total)

        for i in range(start_page - 1, stop):
            page = pdf.pages[i]
            text = page.extract_text()
            if text:
                full_text.append(text)
            if (i - start_page + 2) % 10 == 0:
                print(f"  Extracting page {i + 1}/{stop}...")

    print(f"  Extracted {stop - start_page + 1} pages, {sum(len(t) for t in full_text)} chars")
    return "\n\n".join(full_text)


def extract_txt(filepath: str) -> str:
    with open(filepath, "r", encoding="utf-8") as f:
        return f.read()


def chunk_text(text: str, chunk_size: int, overlap: int) -> list[str]:
    """Semantic chunking: split by paragraphs, merge short ones, split long ones."""
    # Step 1: split into paragraphs
    paragraphs = [p.strip() for p in text.split("\n\n") if p.strip()]

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
    chunks: list[str] = []
    for block in merged:
        if len(block) <= chunk_size * 1.5:
            chunks.append(block)
        else:
            sentences = re.split(r"(?<=[。！？；\n])\s*", block)
            buf = ""
            for s in sentences:
                if len(buf) + len(s) < chunk_size:
                    buf += s
                else:
                    if buf:
                        chunks.append(buf.strip())
                    # sliding window overlap: keep last `overlap` chars
                    if overlap > 0 and len(buf) > overlap:
                        buf = buf[-overlap:] + s
                    else:
                        buf = s
            if buf.strip():
                chunks.append(buf.strip())

    return chunks


def embed_texts(texts: list[str], url: str) -> list[list[float]]:
    resp = requests.post(url, json={"texts": texts})
    resp.raise_for_status()
    return resp.json()["embeddings"]


def chunk_list(items: list, size: int):
    for i in range(0, len(items), size):
        yield items[i:i + size]


def main():
    args = parse_args()
    filepath = args.file
    ext = Path(filepath).suffix.lower()

    print(f"Loading: {filepath}")

    if ext == ".pdf":
        raw_text = extract_pdf(filepath, args.start_page, args.end_page)
    elif ext == ".txt":
        raw_text = extract_txt(filepath)
    else:
        print(f"Unsupported format: {ext}. Use .pdf or .txt")
        sys.exit(1)

    if not raw_text.strip():
        print("No text extracted from document.")
        sys.exit(1)

    print(f"Chunking: target={args.chunk_size} overlap={args.chunk_overlap}")
    chunks = chunk_text(raw_text, args.chunk_size, args.chunk_overlap)

    if not chunks:
        print("No chunks produced.")
        sys.exit(1)

    avg_len = sum(len(c) for c in chunks) / len(chunks)
    print(f"Produced {len(chunks)} chunks (avg {avg_len:.0f} chars)")

    if args.dry_run:
        print("\n--- Preview (first 3 chunks) ---")
        for i, c in enumerate(chunks[:3]):
            print(f"\n[Chunk {i+1}] ({len(c)} chars)")
            print(c[:300] + ("..." if len(c) > 300 else ""))
        return

    print("Embedding and indexing...")
    client = QdrantClient(host=args.qdrant_host, port=args.qdrant_port,
                          grpc_port=args.qdrant_port, prefer_grpc=True)

    indexed = 0
    for batch in chunk_list(chunks, args.batch_size):
        embeddings = embed_texts(batch, args.embed_url)

        points = []
        for text, vec in zip(batch, embeddings):
            points.append(PointStruct(
                id=str(uuid.uuid4()),
                vector=vec,
                payload={
                    "content": text,
                    "book_title": args.book_title,
                    "chapter": args.chapter,
                    "page": 0,
                },
            ))

        client.upsert(collection_name=args.collection, points=points)
        indexed += len(points)
        print(f"  {indexed}/{len(chunks)} chunks indexed")

    count = client.count(collection_name=args.collection)
    print(f"Done. Collection '{args.collection}' now has {count.count} total points.")


if __name__ == "__main__":
    main()
