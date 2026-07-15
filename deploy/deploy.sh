#!/usr/bin/env bash
set -euo pipefail

# ── SkillPass Deploy to VPS ────────────────────────────────
# SSH into VPS, pull latest images, and restart services.
#
# Usage:
#   ./deploy/deploy.sh                          # deploy with 'latest' tag
#   ./deploy/deploy.sh v1.2.3                   # deploy with specific tag
#
# Prerequisites:
#   - SSH access to VPS configured in ~/.ssh/config or via SSH_KEY
#   - DOCKER_USER and VPS_HOST set in deploy/.env
#   - Docker & docker compose installed on VPS
# ───────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Load env
if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
fi

TAG="${1:-latest}"
VPS_HOST="${VPS_HOST:-}"
DOCKER_USER="${DOCKER_USER:-}"

if [ -z "$VPS_HOST" ]; then
  echo "ERROR: VPS_HOST is not set."
  echo "Set it in deploy/.env: VPS_HOST=user@your-vps-ip"
  exit 1
fi

if [ -z "$DOCKER_USER" ]; then
  echo "ERROR: DOCKER_USER is not set."
  exit 1
fi

echo "═══ Deploying SkillPass to $VPS_HOST ═══"
echo "  Tag: $TAG"
echo ""

# ── 1. Ensure remote directory exists ──
echo ">>> Setting up remote directory..."
ssh "$VPS_HOST" "mkdir -p ~/skillpass"

# ── 2. Copy deploy assets to VPS ──
echo ">>> Copying deployment files..."
scp "$SCRIPT_DIR/docker-compose.yml" "$VPS_HOST:~/skillpass/"
scp "$SCRIPT_DIR/.env" "$VPS_HOST:~/skillpass/" 2>/dev/null || \
  echo "  (no .env found locally — ensure .env exists on VPS)"

# ── 3. Secure .env on VPS ──
echo ">>> Securing .env on VPS..."
ssh "$VPS_HOST" "chmod 600 ~/skillpass/.env"

# ── 4. Pull images and restart on VPS ──
echo ">>> Deploying on VPS..."
ssh "$VPS_HOST" "cd ~/skillpass && \
  DOCKER_USER=$DOCKER_USER TAG=$TAG docker compose pull && \
  DOCKER_USER=$DOCKER_USER TAG=$TAG docker compose up -d && \
  echo '✓ Services restarted'"

# ── 4. Show status ──
echo ""
echo ">>> Service status:"
ssh "$VPS_HOST" "cd ~/skillpass && docker compose ps"

echo ""
echo "✓ Deployment complete!"
echo ""
echo "To check logs:"
echo "  ssh $VPS_HOST 'cd ~/skillpass && docker compose logs -f'"
echo ""
echo "To run migrations (first time or after DB changes):"
echo "  ssh $VPS_HOST 'cd ~/skillpass && docker compose --profile setup run --rm migrate'"
echo ""
echo "To seed data (first time only):"
echo "  ssh $VPS_HOST 'cd ~/skillpass && docker compose --profile setup run --rm seed'"
