#!/usr/bin/env bash
set -euo pipefail

# ── SkillPass Docker Build & Push ──────────────────────────
# Builds all Docker images and pushes them to Docker Hub.
#
# Usage:
#   ./deploy/build-and-push.sh                    # build + push with 'latest' tag
#   ./deploy/build-and-push.sh v1.2.3             # build + push with specific tag
#
# Prerequisites:
#   - Docker installed and running
#   - `docker login` completed
#   - DOCKER_USER set in deploy/.env or passed via env
# ───────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load env if available
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
fi

# Config
DOCKER_USER="${DOCKER_USER:-}"
TAG="${1:-latest}"

if [ -z "$DOCKER_USER" ]; then
  echo "ERROR: DOCKER_USER is not set."
  echo "Set it in deploy/.env or export DOCKER_USER=your-username"
  exit 1
fi

echo "═══ Building SkillPass Docker images ═══"
echo "  Registry: docker.io/$DOCKER_USER"
echo "  Tag:      $TAG"
echo ""

# ── 1. Build server image (multi-stage: builds Go binaries + embeds web dist) ──
echo ">>> Building skillpass-server..."
docker build \
  -f "$PROJECT_DIR/server-go/Dockerfile" \
  -t "$DOCKER_USER/skillpass-server:$TAG" \
  "$PROJECT_DIR"

# ── 2. Build markitdown image ──
echo ">>> Building skillpass-markitdown..."
docker build \
  -f "$PROJECT_DIR/markitdown-service/Dockerfile" \
  -t "$DOCKER_USER/skillpass-markitdown:$TAG" \
  "$PROJECT_DIR/markitdown-service"

echo ""
echo "═══ Pushing images to Docker Hub ═══"

docker push "$DOCKER_USER/skillpass-server:$TAG"
docker push "$DOCKER_USER/skillpass-markitdown:$TAG"

echo ""
echo "✓ All images built and pushed successfully!"
echo ""
echo "To deploy on VPS:"
echo "  1. Copy deploy/.env.prod.example to deploy/.env on VPS and fill in values"
echo "  2. Run: ./deploy/deploy.sh $TAG"
