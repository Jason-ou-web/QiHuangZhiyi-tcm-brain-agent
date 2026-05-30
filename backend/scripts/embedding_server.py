"""
Local Embedding + Reranker server for agri-qa-system MVP.
Uses BGE-small-zh-v1.5 for embeddings and BGE-reranker-base for reranking.

Usage:
  # both embedding + reranker
  HF_ENDPOINT=https://hf-mirror.com python embedding_server.py --port 8081

  # embedding only (reranker runs on port 8082, skip for now)
  HF_ENDPOINT=https://hf-mirror.com python embedding_server.py --port 8081 --skip-reranker
"""

import argparse
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer, CrossEncoder
import uvicorn

app = FastAPI(title="agri-qa-embedding-server")

embed_model = None
reranker_model = None
embed_model_name = "none"
reranker_model_name = "none"


class EmbedRequest(BaseModel):
    texts: list[str]


class EmbedResponse(BaseModel):
    embeddings: list[list[float]]


class RerankRequest(BaseModel):
    query: str
    documents: list[str]


class RerankResult(BaseModel):
    index: int
    score: float


class RerankResponse(BaseModel):
    results: list[RerankResult]


@app.post("/embed")
def embed(req: EmbedRequest) -> EmbedResponse:
    vectors = embed_model.encode(req.texts, normalize_embeddings=True)
    return EmbedResponse(embeddings=[v.tolist() for v in vectors])


@app.post("/rerank")
def rerank(req: RerankRequest) -> RerankResponse:
    if reranker_model is None:
        raise HTTPException(status_code=503, detail="reranker not loaded")
    pairs = [(req.query, doc) for doc in req.documents]
    scores = reranker_model.predict(pairs)
    results = [
        RerankResult(index=i, score=float(s))
        for i, s in sorted(enumerate(scores), key=lambda x: -x[1])
    ]
    return RerankResponse(results=results)


@app.get("/health")
def health():
    return {
        "status": "ok",
        "embed_model": embed_model_name,
        "reranker_model": reranker_model_name,
    }


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=8081)
    parser.add_argument("--embed-model", default="BAAI/bge-small-zh-v1.5")
    parser.add_argument("--reranker-model", default="BAAI/bge-reranker-base")
    parser.add_argument("--skip-reranker", action="store_true")
    parser.add_argument("--device", default="cpu")
    args = parser.parse_args()

    embed_model_name = args.embed_model
    reranker_model_name = args.reranker_model if not args.skip_reranker else "disabled"

    print(f"Loading embedding model: {args.embed_model} ...")
    embed_model = SentenceTransformer(args.embed_model, device=args.device)
    print("Embedding model loaded.")

    if args.skip_reranker:
        print("Reranker disabled (--skip-reranker).")
    else:
        print(f"Loading reranker model: {args.reranker_model} ...")
        try:
            reranker_model = CrossEncoder(args.reranker_model, max_length=512)
            print("Reranker model loaded.")
        except Exception as e:
            print(f"WARNING: Failed to load reranker: {e}")
            print("Reranker endpoint will return 503, backend will fallback to vector score.")

    print(f"Starting server on port {args.port}")
    uvicorn.run(app, host="127.0.0.1", port=args.port, log_level="info")
