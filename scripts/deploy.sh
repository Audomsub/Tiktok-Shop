#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/home/audomsub/tiktok-shop"
COMPOSE_FILE="docker-compose.prod.yml"

echo "=========================================================="
echo " Starting TikTok Affiliate Analytics Deployment"
echo " Time: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "=========================================================="

# 1. Navigate to application directory
if [ ! -d "$APP_DIR" ]; then
    echo "Directory $APP_DIR does not exist. Creating and cloning..."
    git clone https://github.com/Audomsub/Tiktok-Shop.git "$APP_DIR"
fi

cd "$APP_DIR"
git config --global --add safe.directory "$APP_DIR"

# 2. Synchronize git repository
echo "==> [1/6] Pulling latest code from origin/main..."
git fetch origin main
git reset --hard origin/main

# 3. Verify .env file
echo "==> [2/6] Verifying environment configuration..."
if [ ! -f .env ]; then
    echo "ERROR: Missing .env file in $APP_DIR"
    exit 1
fi

# 4. Synchronize and reload Nginx configuration
echo "==> [3/6] Applying Nginx configuration..."
sudo cp "$APP_DIR/nginx/tiktok-shop.conf" /etc/nginx/sites-available/default
sudo nginx -t
sudo systemctl reload nginx

# 5. Build and launch Docker containers
echo "==> [4/6] Building and starting production containers..."
docker compose -f "$COMPOSE_FILE" build --parallel
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans

# 6. Healthcheck polling
echo "==> [5/6] Verifying service health..."

# Check Go Backend (internal port 8081)
MAX_RETRIES=20
COUNT=0
echo "Checking Go Backend health (http://127.0.0.1:8081/health)..."
until curl -sSf http://127.0.0.1:8081/health > /dev/null 2>&1; do
    COUNT=$((COUNT + 1))
    if [ $COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Go Backend failed healthcheck after $MAX_RETRIES attempts."
        docker compose -f "$COMPOSE_FILE" logs backend
        exit 1
    fi
    sleep 2
done
echo "Go Backend is HEALTHY!"

# Check Next.js Frontend (internal port 3000)
COUNT=0
echo "Checking Next.js Frontend (http://127.0.0.1:3000)..."
until curl -sSf http://127.0.0.1:3000 > /dev/null 2>&1; do
    COUNT=$((COUNT + 1))
    if [ $COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Next.js Frontend failed to respond after $MAX_RETRIES attempts."
        docker compose -f "$COMPOSE_FILE" logs frontend
        exit 1
    fi
    sleep 2
done
echo "Next.js Frontend is HEALTHY!"

# Check Nginx Gateway Routing (port 80)
COUNT=0
echo "Checking Nginx Gateway (http://127.0.0.1/)..."
until curl -sSf http://127.0.0.1/ > /dev/null 2>&1; do
    COUNT=$((COUNT + 1))
    if [ $COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Nginx Gateway failed to route traffic."
        exit 1
    fi
    sleep 1
done
echo "Nginx Gateway is ROUTING PROPERLY!"

# 7. Cleanup dangling docker resources
echo "==> [6/6] Pruning unused Docker images..."
docker image prune -f

echo "=========================================================="
echo " Deployment Successfully Completed!"
echo " Frontend URL: http://localhost:8080/ (via host port-forward)"
echo " Backend API:  http://localhost:8080/api/v1/ping"
echo " Time: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "=========================================================="
