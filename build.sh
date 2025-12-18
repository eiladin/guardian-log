#!/bin/bash
set -e

echo "🏗️  Building Guardian-Log..."

# Step 1: Build frontend
echo "📦 Building frontend..."
cd web
npm run build
cd ..

# Step 2: Copy dist to webfs for embedding
echo "📁 Copying frontend dist to webfs..."
rm -rf webfs/web
mkdir -p webfs/web
cp -r web/dist webfs/web/

# Step 3: Build Go backend with embedded frontend
echo "🔨 Building Go backend..."
go build -o ./bin/guardian-log ./cmd/guardian-log

echo "✅ Build complete! Binary: ./bin/guardian-log"
echo ""
echo "Run with: ./bin/guardian-log"
echo "Dashboard will be available at: http://localhost:8080"
