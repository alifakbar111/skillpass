"""Lightweight FastAPI wrapper around Microsoft MarkItDown.

Converts uploaded PDF files to Markdown text.  The Go server calls this
service during POST /profiles/me/resume-upload so that the LLM receives
rich Markdown instead of plain text extracted by the basic Go PDF library.

When an LLM is configured (OPENAI_API_KEY), messy Canva-style PDFs are
automatically cleaned up into structured Markdown.

Endpoints
---------
POST /convert   – multipart file upload → { "markdown": "..." }
GET  /health    – liveness probe
"""

from __future__ import annotations

import logging
import os
import tempfile
import uuid
from datetime import datetime, timezone
from pathlib import Path

# Load .env from server-go/ for local dev (LLM keys, etc.)
try:
    from dotenv import load_dotenv

    _env_path = Path(__file__).resolve().parent.parent / "server-go" / ".env"
    if _env_path.exists():
        load_dotenv(_env_path)
        logger_loaded = True
    else:
        logger_loaded = False
except ImportError:
    logger_loaded = False

from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import JSONResponse
from markitdown import MarkItDown

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

if logger_loaded:
    logger.info("loaded env from %s", _env_path)

app = FastAPI(title="MarkItDown Service", version="0.1.0")

# Reuse one MarkItDown instance across requests.
_md = MarkItDown()

MAX_FILE_BYTES = 8 * 1024 * 1024  # 8 MB – matches Go handler limit

# Directory to save exported markdown files.
EXPORT_DIR = Path(__file__).parent / "export-pdf"
EXPORT_DIR.mkdir(exist_ok=True)

# ── LLM client (optional, for cleaning up messy PDFs) ──────────────
_llm_client = None
_llm_model = os.getenv("LLM_MODEL", "gpt-4o-mini")

# Support both standard env vars and the server's LLM_API_KEY convention.
_OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "") or os.getenv("LLM_API_KEY", "")
_ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")
_LLM_PROVIDER = os.getenv("LLM_PROVIDER", "openai").lower()

if _LLM_PROVIDER == "anthropic" and _ANTHROPIC_API_KEY:
    try:
        import anthropic

        _llm_client = anthropic.Anthropic(api_key=_ANTHROPIC_API_KEY)
        _llm_model = os.getenv("LLM_MODEL", "claude-sonnet-4-20250514")
        logger.info("LLM cleanup enabled (Anthropic, model=%s)", _llm_model)
    except ImportError:
        logger.warning("anthropic package not installed — LLM cleanup disabled")
elif _OPENAI_API_KEY:
    try:
        from openai import OpenAI

        _base_url = os.getenv("LLM_BASE_URL", None)
        kwargs = {"api_key": _OPENAI_API_KEY}
        if _base_url:
            kwargs["base_url"] = _base_url
        _llm_client = OpenAI(**kwargs)
        logger.info("LLM cleanup enabled (OpenAI, model=%s, base_url=%s)",
                     _llm_model, _base_url or "default")
    except ImportError:
        logger.warning("openai package not installed — LLM cleanup disabled")
else:
    logger.info("No LLM API key found — raw markdown only (no cleanup)")


def _looks_like_messy_extract(text: str) -> bool:
    """Heuristic: detect messy Canva-style PDF extractions.

    Signs of bad extraction:
    - Very few real newlines relative to length
    - Lots of pipe characters (table artifacts from design tools)
    - Repeated table separator lines (----+----+----)
    """
    if len(text) < 200:
        return False

    # Lots of pipe chars = table artifacts from design tools
    pipe_count = text.count("|")
    if pipe_count > 10:
        return True

    # Repeated table separator patterns
    import re
    if re.search(r"(\|[\s-]+\|[\s-]+\|)", text):
        return True

    # Low newline ratio (pure concatenation)
    newline_ratio = text.count("\n") / len(text)
    return newline_ratio < 0.01


_CLEANUP_PROMPT = """\
You are a resume formatting assistant. The text below was extracted from a \
PDF resume using automated text extraction. The layout may be messy — text \
elements from a design tool (like Canva) were concatenated without proper \
formatting.

Your job: rewrite the text as clean, well-structured Markdown that preserves \
ALL the original content. Use these rules:

- Use # for the person's name
- Use ## for section headings (Education, Experience, Skills, etc.)
- Use bullet points for lists (skills, responsibilities, etc.)
- Use **bold** for job titles and company names
- Use proper line breaks between sections
- Preserve ALL dates, contact info, and details exactly as written
- Do NOT invent or add any content that isn't in the original text
- Do NOT omit any content

Return ONLY the cleaned Markdown, no explanation."""


def _cleanup_with_llm(raw_text: str) -> str | None:
    """Send messy text to the LLM for cleanup. Returns None on failure."""
    if _llm_client is None:
        return None

    try:
        if _LLM_PROVIDER == "anthropic":  # anthropic
            resp = _llm_client.messages.create(
                model=_llm_model,
                max_tokens=4096,
                messages=[{"role": "user", "content": f"{_CLEANUP_PROMPT}\n\n---\n\n{raw_text}"}],
            )
            return resp.content[0].text
        else:  # openai-compatible (OpenAI, 9Router, etc.)
            resp = _llm_client.chat.completions.create(
                model=_llm_model,
                messages=[
                    {"role": "system", "content": _CLEANUP_PROMPT},
                    {"role": "user", "content": raw_text},
                ],
                temperature=0.1,
            )
            return resp.choices[0].message.content
    except Exception as exc:
        logger.warning("LLM cleanup failed: %s", exc)
        return None


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.post("/convert")
async def convert(file: UploadFile = File(...)) -> JSONResponse:
    """Accept a PDF upload and return its Markdown representation."""

    data = await file.read()
    if len(data) > MAX_FILE_BYTES:
        raise HTTPException(status_code=400, detail="File too large (max 8 MB)")

    # MarkItDown needs a file on disk or a file-like object.
    # Writing to a temp file is the safest path since some converters
    # need random access (seek).
    with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as tmp:
        tmp.write(data)
        tmp_path = Path(tmp.name)

    try:
        result = _md.convert(str(tmp_path))
        markdown = result.text_content
    except Exception as exc:
        logger.warning("markitdown conversion failed: %s", exc)
        raise HTTPException(status_code=422, detail="Could not convert PDF to Markdown")
    finally:
        tmp_path.unlink(missing_ok=True)

    # If the extraction looks messy, try LLM cleanup.
    cleaned = None
    if _looks_like_messy_extract(markdown):
        logger.info("messy PDF detected (%d chars, %.3f newline ratio), running LLM cleanup",
                     len(markdown), markdown.count("\n") / max(len(markdown), 1))
        cleaned = _cleanup_with_llm(markdown)

    final_markdown = cleaned if cleaned else markdown

    # Save a copy with a unique filename: <original-stem>_<YYYYMMDD>_<short-uuid>.md
    stem = Path(file.filename or "resume").stem
    date_part = datetime.now(timezone.utc).strftime("%Y%m%d")
    short_id = uuid.uuid4().hex[:8]
    export_name = f"{stem}_{date_part}_{short_id}.md"
    export_path = EXPORT_DIR / export_name

    try:
        export_path.write_text(final_markdown, encoding="utf-8")
        logger.info("saved markdown export → %s", export_path)
    except Exception as exc:
        logger.warning("failed to save export: %s", exc)

    return JSONResponse(content={"markdown": final_markdown, "exportedTo": export_name})
