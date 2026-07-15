# SkillPass — Dev Server Deployment

Deploy SkillPass to a VPS using Docker containers from Docker Hub.

## Architecture

```
┌─ Internet ──────────────────────────────────────────────┐
│  http://<VPS_IP>                                         │
│         │                                                │
│         ▼                                                │
│    ┌──────────┐     (proxy to :1234)                     │
│    │  Nginx   │ ──────────────────► ┌──────────────┐    │
│    │ (host)   │                     │  Go Server   │    │
│    │ port 80  │                     │  (API + SPA) │    │
│    └──────────┘                     │  port 1234   │    │
│                                    └──────┬───────┘    │
│                        ┌──────────────────┼─────┐       │
│                        ▼                  ▼     ▼       │
│                   ┌──────────┐     ┌──────────────┐     │
│                   │PostgreSQL│     │  MarkItDown  │     │
│                   │ port 5432│     │  port 8000   │     │
│                   └──────────┘     └──────────────┘     │
└─────────────────────────────────────────────────────────┘
```

**No domain needed:** Access by VPS IP address (e.g., `http://203.0.113.42`)

## Prerequisites (Local Machine)

- Docker installed and running
- `docker login` (Docker Hub account)
- SSH access to VPS

## Prerequisites (VPS - One-Time Setup)

SSH into your VPS and run the setup script:

```bash
# Option 1: Pipe directly
ssh root@<VPS_IP> 'bash -s' < deploy/setup-vps.sh

# Option 2: Copy and run
scp deploy/setup-vps.sh root@<VPS_IP>:~
ssh root@<VPS_IP> ./setup-vps.sh
```

This installs: Docker, Docker Compose, Nginx, and configures the firewall.

Then set up Nginx reverse proxy:

```bash
ssh root@<VPS_IP>

# Copy nginx config
sudo cp ~/skillpass/nginx.conf /etc/nginx/sites-available/skillpass
sudo ln -s /etc/nginx/sites-available/skillpass /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

(You'll need to copy the nginx config after the first deploy or scp it separately.)

## Deploy Flow (Do this from your local machine)

### 1. Configure Environment

```bash
cp deploy/.env.prod.example deploy/.env
# Edit deploy/.env with your values:
#   VPS_HOST=root@<VPS_IP>
#   DOCKER_USER=your-dockerhub-username
#   JWT_SECRET=<random-strong-secret>  (generate with: openssl rand -base64 32)
#   CORS_ORIGIN=http://<VPS_IP>
#   APP_URL=http://<VPS_IP>
#
# DATABASE_URL is auto-set by docker-compose to point at the local db container
```

### 2. Build & Push Images

```bash
./deploy/build-and-push.sh
```

This builds and pushes to Docker Hub:
- `your-user/skillpass-server:latest`
- `your-user/skillpass-web:latest`
- `your-user/skillpass-markitdown:latest`

### 3. Deploy to VPS

```bash
./deploy/deploy.sh
```

This SSHes into the VPS, pulls images, and starts containers.

### 4. Run Migrations (first time only)

```bash
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose --profile setup run --rm migrate'
```

### 5. Seed Data (optional, first time only)

```bash
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose --profile setup run --rm seed'
```

## Regular Updates

When you have new code to deploy:

```bash
# 1. Build and push new images
./deploy/build-and-push.sh

# 2. Deploy (pulls new images and restarts)
./deploy/deploy.sh
```

## Useful Commands

```bash
# View logs
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose logs -f'

# View logs for specific service
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose logs -f server'

# Restart services
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose restart'

# Stop everything
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose down'

# Check service status
ssh root@<VPS_IP> 'cd ~/skillpass && docker compose ps'
```

## Rollback

To roll back to a previous version:

```bash
# Deploy a specific tag
./deploy/deploy.sh v1.0.0
```

Or manually on the VPS:

```bash
ssh root@<VPS_IP>
cd ~/skillpass
TAG=previous-tag docker compose up -d
```

## Security Notes

- Change `JWT_SECRET` to a strong random value (use `openssl rand -base64 32`)
- Port 1234 only listens on `127.0.0.1` — not exposed to the internet
- All external traffic goes through Nginx on port 80
- SSL/TLS: Add a domain and use Certbot / Caddy for HTTPS when ready
- The `.env` file on VPS contains secrets — protect it (`chmod 600`)

## Files in deploy/

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Docker Compose config for VPS |
| `.env.prod.example` | Environment variable template |
| `nginx.conf` | Nginx reverse proxy config (for VPS host) |
| `build-and-push.sh` | Build images + push to Docker Hub |
| `deploy.sh` | SSH into VPS and deploy |
| `setup-vps.sh` | One-time VPS provisioning script |
