#!/usr/bin/env bash
# Copy this to deploy.sh (gitignored) and fill in your actual values.
set -euo pipefail

# ── Config ───────────────────────────────────────────────────────────
SERVER="user@YOUR_SERVER_IP"
DEPLOY_USER="deploy"
REMOTE_DIR="/home/deploy/app"
DOMAIN="your.domain.com"
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

# ── Step 2: Package ──────────────────────────────────────────────────
echo "==> Packaging deployment..."
tar czf /tmp/speed-deploy.tar.gz \
  bin/api-linux \
  bin/scraper-linux \
  frontend/build \
  data/geojson \
  migrations \
  .env.example

# ── Step 3: Upload ───────────────────────────────────────────────────
echo "==> Splitting and uploading..."
split -b "$CHUNK_SIZE" /tmp/speed-deploy.tar.gz /tmp/speed-chunk-
ssh "$SERVER" "mkdir -p /tmp/speed-chunks"
for chunk in /tmp/speed-chunk-*; do
  scp -q "$chunk" "$SERVER:/tmp/speed-chunks/$(basename "$chunk")"
done

# ── Step 4: Deploy on server ─────────────────────────────────────────
ssh "$SERVER" bash -s << REMOTE
set -e
cat /tmp/speed-chunks/speed-chunk-* > /tmp/speed-deploy.tar.gz
rm -rf /tmp/speed-chunks
mkdir -p $REMOTE_DIR
cd $REMOTE_DIR
tar xzf /tmp/speed-deploy.tar.gz
rm /tmp/speed-deploy.tar.gz
mv bin/api-linux bin/api
mv bin/scraper-linux bin/scraper
chmod +x bin/api bin/scraper
chown -R $DEPLOY_USER:$DEPLOY_USER $REMOTE_DIR

# Create .env if missing — fill in actual credentials
if [ ! -f .env ]; then
  cat > .env << 'ENV'
DATABASE_URL=postgres://YOUR_DB_USER:YOUR_DB_PASSWORD@localhost:5432/speed_limit_ke?sslmode=disable
PORT=8080
GIN_MODE=release
ENV
fi

# Run migrations (idempotent)
sudo -u $DEPLOY_USER psql "\$(grep DATABASE_URL .env | cut -d= -f2-)" \
  -f migrations/001_create_road_segments.up.sql 2>/dev/null || true
sudo -u $DEPLOY_USER psql "\$(grep DATABASE_URL .env | cut -d= -f2-)" \
  -f migrations/002_create_hazards.up.sql 2>/dev/null || true

# Seed data
sudo -u $DEPLOY_USER bash -c "cd $REMOTE_DIR && ./bin/scraper --seed" 2>/dev/null || true

systemctl daemon-reload
systemctl restart speed-api
systemctl reload caddy
echo "=== Deployment complete ==="
REMOTE

rm -f /tmp/speed-deploy.tar.gz /tmp/speed-chunk-*
echo "==> Done! Site: https://$DOMAIN"
