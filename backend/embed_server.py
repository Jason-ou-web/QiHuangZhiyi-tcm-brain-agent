"""
TCM Embedding & Reranker Server
Serves OpenAI-compatible /embeddings and /rerank on a single port.
Usage:
  python embed_server.py                  # both services on port 8081
  python embed_server.py --port 8081       # custom port
  python embed_server.py --embed-only      # skip reranker
  python embed_server.py --rerank-only     # skip embedding
"""

import argparse
import os
import sys

import uvicorn
from fastapi import FastAPI
from pydantic import BaseModel

# ---------- config (override via env) ----------
EMBED_MODEL = os.getenv("EMBED_MODEL", "BAAI/bge-small-zh-v1.5")
RERANK_MODEL = os.getenv("RERANK_MODEL", "BAAI/bge-reranker-base")


# ---------- lazy-load models ----------
_embed_model = None
_rerank_model = None


def get_embed_model():
    global _embed_model
    if _embed_model is None:
        from sentence_transformers import SentenceTransformer

        print(f"[embed] loading {EMBED_MODEL} ...")
        _embed_model = SentenceTransformer(EMBED_MODEL)
        print("[embed] ready")
    return _embed_model


def get_rerank_model():
    global _rerank_model
    if _rerank_model is None:
        from sentence_transformers import CrossEncoder

        print(f"[rerank] loading {RERANK_MODEL} ...")
        _rerank_model = CrossEncoder(RERANK_MODEL)
        print("[rerank] ready")
    return _rerank_model


# ---------- request / response ----------
class EmbedRequest(BaseModel):
    input: list[str]
    model: str = ""


class EmbedData(BaseModel):
    embedding: list[float]
    index: int


class EmbedResponse(BaseModel):
    data: list[EmbedData]


class RerankRequest(BaseModel):
    query: str
    documents: list[str]
    model: str = ""


class RerankData(BaseModel):
    index: int
    score: float


class RerankResponse(BaseModel):
    data: list[RerankData]


# ---------- app ----------
app = FastAPI(title="tcm-embedding-server")


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.post("/embeddings", response_model=EmbedResponse)
async def embeddings(req: EmbedRequest):
    model = get_embed_model()
    vectors = model.encode(req.input, normalize_embeddings=True)
    data = [EmbedData(embedding=v.tolist(), index=i) for i, v in enumerate(vectors)]
    return EmbedResponse(data=data)


@app.post("/rerank", response_model=RerankResponse)
async def rerank(req: RerankRequest):
    model = get_rerank_model()
    pairs = [(req.query, doc) for doc in req.documents]
    scores = model.predict(pairs)
    data = sorted(
        [RerankData(index=i, score=float(s)) for i, s in enumerate(scores)],
        key=lambda x: -x.score,
    )
    return RerankResponse(data=data)


# ---------- main ----------
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="TCM Embedding Server")
    parser.add_argument("--port", type=int, default=8081)
    parser.add_argument("--embed-only", action="store_true")
    parser.add_argument("--rerank-only", action="store_true")
    args = parser.parse_args()

    if args.embed_only and args.rerank_only:
        print("Error: cannot use both --embed-only and --rerank-only")
        sys.exit(1)

    if args.embed_only:
        # remove rerank route to avoid 500 on accidental call
        app.routes = [r for r in app.routes if r.path != "/rerank"]
    elif args.rerank_only:
        app.routes = [r for r in app.routes if r.path != "/embeddings"]

    print(f"Server: http://127.0.0.1:{args.port}")
    if not args.rerank_only:
        print(f"  POST /embeddings  <- {EMBED_MODEL}")
    if not args.embed_only:
        print(f"  POST /rerank      <- {RERANK_MODEL}")

    uvicorn.run(app, host="127.0.0.1", port=args.port, log_level="info")
