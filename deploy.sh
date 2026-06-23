#!/usr/bin/env bash
set -euo pipefail

# ── Config ───────────────────────────────────────────────────────────
SERVER="root@164.92.173.56"
DEPLOY_USER="deploy"
REMOTE_DIR="/home/deploy/app"
DOMAIN="speed.koru.africa"
CHUNK_SIZE="50m"

# ── Step 1: Build locally ────────────────────────────────────────────
echo "==> Building Go API (linux/amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-linux cmd/api/main.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/scraper-linux cmd/scraper/main.go

echo "==> Building frontend..."
cd frontend
pnpm install --frozen-lockfile
pnpm run build
cd ..

# ── Step 2: Create tarball and split ─────────────────────────────────
echo "==> Packaging deployment..."
tar czf /tmp/speed-deploy.tar.gz \
  bin/api-linux \
  bin/scraper-linux \
  frontend/build \
  data/geojson \
  migrations \
  .env.example

FILESIZE=$(stat -c%s /tmp/speed-deploy.tar.gz 2>/dev/null || stat -f%z /tmp/speed-deploy.tar.gz)
echo "    Archive size: $((FILESIZE / 1024 / 1024))MB"

echo "==> Splitting archive..."
split -b "$CHUNK_SIZE" /tmp/speed-deploy.tar.gz /tmp/speed-chunk-

# ── Step 3: Upload chunks ────────────────────────────────────────────
echo "==> Uploading chunks to server..."
ssh "$SERVER" "mkdir -p /tmp/speed-chunks"

for chunk in /tmp/speed-chunk-*; do
  name=$(basename "$chunk")
  echo "    Uploading $name ($(stat -c%s "$chunk" 2>/dev/null || stat -f%z "$chunk") bytes)..."
  scp -q "$chunk" "$SERVER:/tmp/speed-chunks/$name"
done

# ── Step 4: Reassemble and deploy on server ──────────────────────────
echo "==> Deploying on server..."
ssh "$SERVER" bash -s << REMOTE
set -e

# Reassemble
cat /tmp/speed-chunks/speed-chunk-* > /tmp/speed-deploy.tar.gz
rm -rf /tmp/speed-chunks

# Extract
mkdir -p $REMOTE_DIR
cd $REMOTE_DIR
tar xzf /tmp/speed-deploy.tar.gz
rm /tmp/speed-deploy.tar.gz

# Set permissions
mv bin/api-linux bin/api
mv bin/scraper-linux bin/scraper
chmod +x bin/api bin/scraper
chown -R $DEPLOY_USER:$DEPLOY_USER $REMOTE_DIR

# Create .env if missing
if [ ! -f .env ]; then
  cat > .env << 'ENV'
DATABASE_URL=postgres://deploy:speedke2026@localhost:5432/speed_limit_ke?sslmode=disable
PORT=8080
GIN_MODE=release
ENV
  chown $DEPLOY_USER:$DEPLOY_USER .env
fi

# Run migration
sudo -u $DEPLOY_USER psql "postgres://deploy:speedke2026@localhost:5432/speed_limit_ke" \
  -f migrations/001_create_road_segments.up.sql 2>/dev/null || true

# Seed data
sudo -u $DEPLOY_USER bash -c 'cd $REMOTE_DIR && ./bin/scraper --seed' 2>/dev/null || true

# ── Systemd service for Go API ──
cat > /etc/systemd/system/speed-api.service << 'SVC'
[Unit]
Description=Kenya Speed Limits API
After=network.target postgresql.service

[Service]
User=deploy
Group=deploy
WorkingDirectory=/home/deploy/app
EnvironmentFile=/home/deploy/app/.env
ExecStart=/home/deploy/app/bin/api serve
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVC

systemctl daemon-reload
systemctl enable speed-api
systemctl restart speed-api

# ── Caddy config ──
cat > /etc/caddy/Caddyfile << 'CADDY'
$DOMAIN {
    handle /api/* {
        reverse_proxy localhost:8080
    }

    handle /health {
        reverse_proxy localhost:8080
    }

    handle {
        root * /home/deploy/app/frontend/build
        try_files {path} /index.html
        file_server
    }
}
CADDY

systemctl reload caddy

echo "=== Deployment complete ==="
echo "API: systemctl status speed-api"
echo "Site: https://$DOMAIN"
REMOTE

# ── Cleanup local temp files ─────────────────────────────────────────
rm -f /tmp/speed-deploy.tar.gz /tmp/speed-chunk-*

echo ""
echo "==> Done! Site will be live at https://$DOMAIN"
echo "    (Make sure DNS A record for $DOMAIN points to 164.92.173.56)"
