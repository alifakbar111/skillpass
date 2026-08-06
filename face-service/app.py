"""SkillPass face-service — STUB implementation.

Deterministic, dependency-light stand-in for a real ArcFace (embeddings) +
FASNet (liveness) pipeline. The HTTP contract is identical to the production
service, so it can be swapped for real ONNX models without touching the Go
backend or the frontend.

  POST /enroll  { image: base64 }              -> { embedding: base64, livenessScore }
  POST /verify  { image: base64, embedding }   -> { matchScore, livenessScore, passed }

Behaviour: the embedding is a deterministic normalized vector derived from the
image bytes, so re-submitting the same frame verifies with a high match score
and a different image scores low — believable for demos and tests.
"""

import base64
import hashlib
import math
import struct

from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="SkillPass Face Service (stub)")

EMB_DIM = 64
MATCH_THRESHOLD = 0.82
LIVENESS_THRESHOLD = 0.70


def pseudo_embedding(data: bytes) -> list[float]:
    """Deterministic normalized vector from the image bytes."""
    vec: list[float] = []
    seed = data or b"\x00"
    while len(vec) < EMB_DIM:
        seed = hashlib.sha256(seed).digest()
        for i in range(0, len(seed), 4):
            if len(vec) >= EMB_DIM:
                break
            v = struct.unpack(">I", seed[i : i + 4])[0]
            vec.append((v % 2000) / 1000.0 - 1.0)  # -1.0 .. 1.0
    norm = math.sqrt(sum(x * x for x in vec)) or 1.0
    return [x / norm for x in vec]


def pack(vec: list[float]) -> str:
    return base64.b64encode(struct.pack(f">{len(vec)}f", *vec)).decode()


def unpack(b: str) -> list[float]:
    raw = base64.b64decode(b)
    n = len(raw) // 4
    return list(struct.unpack(f">{n}f", raw))


def liveness(data: bytes) -> float:
    """Deterministic 0.85..0.99 liveness score."""
    h = hashlib.sha256(data or b"\x00").digest()[0]
    return round(0.85 + (h % 15) / 100.0, 3)


class EnrollReq(BaseModel):
    image: str


class VerifyReq(BaseModel):
    image: str
    embedding: str


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/enroll")
def enroll(req: EnrollReq):
    data = base64.b64decode(req.image)
    return {"embedding": pack(pseudo_embedding(data)), "livenessScore": liveness(data)}


@app.post("/verify")
def verify(req: VerifyReq):
    data = base64.b64decode(req.image)
    new = pseudo_embedding(data)
    stored = unpack(req.embedding)
    dot = sum(a * b for a, b in zip(new, stored))  # cosine sim (both normalized)
    match = round(max(0.0, min(1.0, dot)), 3)
    lv = liveness(data)
    return {
        "matchScore": match,
        "livenessScore": lv,
        "passed": match >= MATCH_THRESHOLD and lv >= LIVENESS_THRESHOLD,
    }
