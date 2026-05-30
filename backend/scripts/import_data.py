"""
Batch import agricultural knowledge into Qdrant.

Supported formats:
  - JSON: [{"content": "...", "book_title": "...", "chapter": "...", "page": 1}, ...]
  - YAML: same structure as JSON
  - Markdown: # Book Title / ## Chapter / content under each heading

Usage:
  python import_data.py --file data.json
  python import_data.py --file data.yaml
  python import_data.py --file book.md --book-title "齐民要术"
  python import_data.py --file book.md --book-title "齐民要术" --chapter "卷一"
"""

import argparse
import json
import re
import sys
import uuid
from pathlib import Path

import requests
from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct


def parse_args():
    p = argparse.ArgumentParser(description="Import agricultural knowledge into Qdrant")
    p.add_argument("--file", required=True, help="Path to data file (.json/.yaml/.md/.txt)")
    p.add_argument("--book-title", default="", help="Book title (for Markdown, overrides # heading)")
    p.add_argument("--chapter", default="", help="Chapter name (for Markdown, overrides ## heading)")
    p.add_argument("--embed-url", default="http://127.0.0.1:8081/embed")
    p.add_argument("--qdrant-host", default="localhost")
    p.add_argument("--qdrant-port", type=int, default=6334)
    p.add_argument("--collection", default="agri_knowledge")
    p.add_argument("--batch-size", type=int, default=20)
    return p.parse_args()


def embed_texts(texts: list[str], url: str) -> list[list[float]]:
    resp = requests.post(url, json={"texts": texts})
    resp.raise_for_status()
    return resp.json()["embeddings"]


def parse_json(filepath: str) -> list[dict]:
    with open(filepath, "r", encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, list):
        print("Error: JSON must be a list of objects")
        sys.exit(1)
    return data


def parse_yaml(filepath: str) -> list[dict]:
    import yaml
    with open(filepath, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)
    if not isinstance(data, list):
        print("Error: YAML must be a list of objects")
        sys.exit(1)
    return data


def parse_markdown(filepath: str, default_book: str, default_chapter: str) -> list[dict]:
    """Parse Markdown: # = book title, ## = chapter, content = chunks."""
    with open(filepath, "r", encoding="utf-8") as f:
        text = f.read()

    book_title = default_book
    chapter = default_chapter
    current_content: list[str] = []
    chunks: list[dict] = []

    def flush():
        nonlocal current_content
        body = "\n".join(current_content).strip()
        if body:
            chunks.append({
                "content": body,
                "book_title": book_title or "",
                "chapter": chapter or "",
                "page": 0,
            })
        current_content = []

    for line in text.split("\n"):
        if line.startswith("# ") and not line.startswith("## "):
            flush()
            if not default_book:
                book_title = line[2:].strip()
        elif line.startswith("## "):
            flush()
            if not default_chapter:
                chapter = line[3:].strip()
        elif line.startswith("---"):
            flush()
        else:
            current_content.append(line)

    flush()
    return chunks


def chunk_list(items: list, size: int):
    for i in range(0, len(items), size):
        yield items[i:i + size]


def main():
    args = parse_args()
    filepath = args.file
    ext = Path(filepath).suffix.lower()

    print(f"Loading: {filepath}")

    if ext == ".json":
        records = parse_json(filepath)
    elif ext in (".yaml", ".yml"):
        records = parse_yaml(filepath)
    elif ext in (".md", ".markdown", ".txt"):
        records = parse_markdown(filepath, args.book_title, args.chapter)
    else:
        print(f"Unsupported format: {ext}")
        sys.exit(1)

    # validate
    valid = []
    for r in records:
        content = (r.get("content") or "").strip()
        if not content:
            continue
        valid.append({
            "content": content,
            "book_title": r.get("book_title") or args.book_title or "",
            "chapter": r.get("chapter") or args.chapter or "",
            "page": r.get("page", 0),
        })

    if not valid:
        print("No valid records found.")
        sys.exit(1)

    print(f"Parsed {len(valid)} chunks, embedding and indexing...")

    client = QdrantClient(host=args.qdrant_host, port=args.qdrant_port,
                          grpc_port=args.qdrant_port, prefer_grpc=True)

    indexed = 0
    for batch in chunk_list(valid, args.batch_size):
        texts = [r["content"] for r in batch]
        embeddings = embed_texts(texts, args.embed_url)

        points = []
        for rec, vec in zip(batch, embeddings):
            points.append(PointStruct(
                id=str(uuid.uuid4()),
                vector=vec,
                payload={
                    "content": rec["content"],
                    "book_title": rec["book_title"],
                    "chapter": rec["chapter"],
                    "page": rec["page"],
                },
            ))

        client.upsert(collection_name=args.collection, points=points)
        indexed += len(points)
        print(f"  {indexed}/{len(valid)} chunks indexed")

    count = client.count(collection_name=args.collection)
    print(f"Done. Collection '{args.collection}' now has {count.count} total points.")


if __name__ == "__main__":
    main()
