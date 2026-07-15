#!/usr/bin/env bash
set -euo pipefail

# ── SkillPass VPS Setup — one-time provisioning ────────────
# Run ONCE on a fresh VPS to install Docker and configure the server.
# SSH into your VPS first, then run this script.
#
# Usage:
#   ssh root@<VPS_IP> 'bash -s' < deploy/setup-vps.sh
#
# Or copy to VPS and run:
#   scp deploy/setup-vps.sh root@<VPS_IP>:~
#   ssh root@<VPS_IP> ./setup-vps.sh
# ───────────────────────────────────────────────────────────

set -e

echo "═══ SkillPass VPS Setup ═══"

# ── Update system ──────────────────────────────────────────
echo ">>> Updating system packages..."
apt-get update -qq
apt-get upgrade -y -qq

# ── Install Docker ─────────────────────────────────────────
if ! command -v docker &>/dev/null; then
  echo ">>> Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
else
  echo "✓ Docker already installed"
fi

# ── Install Docker Compose plugin ──────────────────────────
if ! docker compose version &>/dev/null; then
  echo ">>> Installing Docker Compose plugin..."
  apt-get install -y -qq docker-compose-plugin
else
  echo "✓ Docker Compose already installed"
fi

# ── Install Nginx ──────────────────────────────────────────
if ! command -v nginx &>/dev/null; then
  echo ">>> Installing Nginx..."
  apt-get install -y -qq nginx
  systemctl enable nginx
  systemctl start nginx
else
  echo "✓ Nginx already installed"
fi

# ── Configure firewall ─────────────────────────────────────
echo ">>> Configuring firewall..."
if command -v ufw &>/dev/null; then
  ufw allow 22/tcp   # SSH
  ufw allow 80/tcp   # HTTP
  ufw allow 443/tcp  # HTTPS (for future)
  ufw --force enable
  echo "✓ Firewall configured"
else
  echo "  (ufw not available — ensure ports 22 and 80 are open)"
fi

# ── Install fail2ban (SSH brute-force protection) ──────────
echo ">>> Installing fail2ban..."
apt-get install -y -qq fail2ban
cat > /etc/fail2ban/jail.local << 'EOFF2B'
[sshd]
enabled = true
maxretry = 5
bantime = 3600
findtime = 600
EOFF2B
systemctl enable fail2ban
systemctl start fail2ban
echo "✓ fail2ban configured (5 retries → 1h ban)"

# ── Harden SSH (key-only auth) ────────────────────────────
echo ">>> Hardening SSH configuration..."
sed -i 's/^#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/^PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/^#PermitRootLogin yes/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
sed -i 's/^PermitRootLogin yes/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
systemctl reload sshd
echo "✓ SSH hardened (password auth disabled, root login: key-only)"

# ── Create app directory ───────────────────────────────────
mkdir -p ~/skillpass

echo ""
echo "═══ VPS Setup Complete ═══"
echo ""
echo "Next steps:"
echo "  1. Copy deploy/.env to ~/skillpass/.env and edit:"
echo "     DATABASE_URL, JWT_SECRET, CORS_ORIGIN, APP_URL"
echo "  2. Copy deploy/nginx.conf to /etc/nginx/sites-available/skillpass"
echo "  3. Enable the site and reload nginx"
echo "  4. From your local machine, run: ./deploy/build-and-push.sh && ./deploy/deploy.sh"
echo ""
echo "VPS IP: $(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')"
