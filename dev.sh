#!/bin/bash
# Simple development server launcher
# Starts both backend and frontend in background

set -e

echo "🚀 Starting Guardian Log Development Servers..."
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  Warning: .env file not found${NC}"
    echo "Copy .env.example to .env and configure it"
    echo "cp .env.example .env"
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping servers..."
    jobs -p | xargs -r kill 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Check if ports are available
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${YELLOW}⚠️  Port 8080 is already in use${NC}"
    echo "Kill the process with: kill \$(lsof -ti:8080)"
    exit 1
fi

if lsof -Pi :5173 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${YELLOW}⚠️  Port 5173 is already in use${NC}"
    echo "Kill the process with: kill \$(lsof -ti:5173)"
    exit 1
fi

# Create logs directory
mkdir -p logs

# Start backend
echo -e "${GREEN}📦 Starting backend (port 8080)...${NC}"
if command -v air >/dev/null 2>&1; then
    air > logs/backend.log 2>&1 &
    BACKEND_PID=$!
    echo "   Backend PID: $BACKEND_PID (using air hot-reload)"
else
    echo -e "${YELLOW}   air not found, using go run (no hot-reload)${NC}"
    go run ./cmd/guardian-log > logs/backend.log 2>&1 &
    BACKEND_PID=$!
    echo "   Backend PID: $BACKEND_PID"
fi

# Wait for backend to start
echo "   Waiting for backend to start..."
sleep 3

# Check if backend is running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${YELLOW}❌ Backend failed to start. Check logs/backend.log${NC}"
    tail -20 logs/backend.log
    exit 1
fi

# Start frontend
echo -e "${GREEN}🎨 Starting frontend (port 5173)...${NC}"
cd web && npm run dev > ../logs/frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..
echo "   Frontend PID: $FRONTEND_PID"

# Wait for frontend to start
echo "   Waiting for frontend to start..."
sleep 3

# Check if frontend is running
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${YELLOW}❌ Frontend failed to start. Check logs/frontend.log${NC}"
    tail -20 logs/frontend.log
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Both servers are running!${NC}"
echo ""
echo "📱 Frontend:  http://localhost:5173  (use this for development)"
echo "🔧 Backend:   http://localhost:8080"
echo "💚 Health:    http://localhost:8080/api/health"
echo ""
echo "📋 Logs:"
echo "   Backend:  tail -f logs/backend.log"
echo "   Frontend: tail -f logs/frontend.log"
echo ""
echo "Press Ctrl+C to stop both servers"
echo ""

# Wait for both processes
wait
