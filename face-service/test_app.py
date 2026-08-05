import base64

from fastapi.testclient import TestClient

from app import app

client = TestClient(app)


def b64(s: bytes) -> str:
    return base64.b64encode(s).decode()


def test_health():
    assert client.get("/health").json() == {"status": "ok"}


def test_enroll_returns_embedding_and_liveness():
    r = client.post("/enroll", json={"image": b64(b"face-of-alice")})
    body = r.json()
    assert r.status_code == 200
    assert body["embedding"]
    assert 0.85 <= body["livenessScore"] <= 0.99


def test_same_face_verifies_high():
    img = b64(b"face-of-alice")
    emb = client.post("/enroll", json={"image": img}).json()["embedding"]
    v = client.post("/verify", json={"image": img, "embedding": emb}).json()
    assert v["matchScore"] >= 0.99  # identical frame -> near-perfect match
    assert v["passed"] is True


def test_different_face_verifies_low():
    emb = client.post("/enroll", json={"image": b64(b"face-of-alice")}).json()["embedding"]
    v = client.post("/verify", json={"image": b64(b"face-of-bob-different"), "embedding": emb}).json()
    assert v["matchScore"] < 0.82
    assert v["passed"] is False
