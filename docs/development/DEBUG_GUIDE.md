# VS Code Debugging Guide

## Quick Start - It's Easier Than You Think!

### Simple 3-Step Process

1. **Press F5** in VS Code
   - Select your debug configuration
   - Servers start automatically!

2. **Open browser manually**
   - Navigate to http://localhost:5173 when ready
   - No auto-launch, you control when

3. **Debug!**
   - Set breakpoints by clicking line numbers
   - Use the app to trigger your code
   - Inspect, step through, fix issues

That's it! No manual terminal commands, no separate server startup.

## Available Debug Configurations

### Debug Backend
- **What it does:** Starts frontend dev server + launches Go debugger for backend
- **When to use:** Debugging API handlers, business logic, database operations
- **Auto-starts:** Frontend dev server (port 5173)
- **Debugs:** Backend Go code
- **Breakpoints:** Set in `.go` files
- **Browser:** Open http://localhost:5173 manually when ready

### Debug Frontend
- **What it does:** Starts frontend dev server with Node debugger
- **When to use:** Debugging React components, build process, Vite config
- **Auto-starts:** Frontend dev server only
- **Debugs:** Frontend build process
- **Breakpoints:** Set in `.tsx`, `.ts`, `.jsx`, `.js` files
- **Browser:** Open http://localhost:5173 manually when ready
- **Note:** For React debugging, use browser DevTools (F12) instead

### Debug Full Stack
- **What it does:** Starts frontend dev server + launches Go debugger (same as Debug Backend)
- **When to use:** Debugging API calls end-to-end, full application flow
- **Auto-starts:** Frontend dev server + backend
- **Debugs:** Backend Go code (use browser DevTools for frontend)
- **Breakpoints:** Set in `.go` files
- **Browser:** Open http://localhost:5173 manually when ready

### Attach to Backend
- **What it does:** Attaches to a running Delve debug session
- **When to use:** Advanced debugging with Delve CLI
- **Requires:** Manual Delve startup
- **Setup:**
  ```bash
  dlv debug ./cmd/guardian-log --headless --listen=:2345 --api-version=2
  ```

## Debugging Workflow Examples

### Example 1: Debug API Endpoint

**Scenario:** API call from frontend to backend failing

1. **Set breakpoints:**
   - Open `internal/api/handlers.go`
   - Click line number in handler function
   - Breakpoint set (red dot appears)

2. **Press F5** → Select **"Debug Full Stack"**
   - Frontend dev server starts automatically
   - Backend starts in debug mode
   - Watch terminal panel for "ready" messages

3. **Open browser manually:**
   - Navigate to http://localhost:5173
   - Use the app normally

4. **Trigger the API call:**
   - Click button, perform action
   - Debugger stops at your breakpoint!

5. **Debug:**
   - Inspect variables (hover or watch panel)
   - Step through code (F10, F11)
   - Check call stack
   - Continue (F5) or stop (Shift+F5)

### Example 2: Debug React Component

**Scenario:** Component not rendering correctly

1. **Open browser DevTools** (recommended for React):
   - Press F5 → **"Debug Frontend"** (starts dev server)
   - Open http://localhost:5173 manually
   - Press F12 in browser
   - Go to Sources tab
   - Find your component file
   - Set breakpoints in browser

2. **Or use console.log** (faster for React):
   ```typescript
   console.log('Props:', props);
   console.table(anomalies);
   ```
   - Edit `web/src/components/AnomalyCard.tsx`
   - Save → instant hot reload
   - Check browser console (F12)

3. **React DevTools** (best for component inspection):
   - Install React Developer Tools extension
   - F12 → React tab
   - Inspect component tree
   - View props/state in real-time
   - No breakpoints needed!

### Example 3: Debug Backend Logic Only

**Scenario:** Issue in domain detection logic

1. **Press F5** → **"Debug Backend"**
   - Servers start automatically
   - Wait for "ready" message in terminal

2. **Set breakpoints:**
   - Open `internal/analyzer/detector.go`
   - Click line numbers in functions
   - Red dots appear

3. **Open browser:**
   - Go to http://localhost:5173
   - Use app to trigger the logic

4. **Debug hits breakpoint:**
   - Inspect variables (hover or watch)
   - Step through algorithm (F10/F11)
   - Evaluate expressions in Debug Console
   - Fix the issue!

## Debugging Tips

### Backend Debugging

**Inspect Variables:**
- Hover over variables to see values
- Add to Watch panel (right-click → Add to Watch)
- Use Debug Console to evaluate expressions

**Call Stack:**
- See call hierarchy in Call Stack panel
- Click frames to see context at each level

**Conditional Breakpoints:**
- Right-click breakpoint → Edit Breakpoint
- Add condition: `domain == "example.com"`
- Only stops when condition is true

### Frontend Debugging

**React Component State:**
- Install React Developer Tools extension
- View component tree and state
- Time-travel debugging

**Console Logging:**
- Still useful alongside debugger
- Use `console.log()`, `console.table()`
- See output in Debug Console

**Network Inspection:**
- Open browser DevTools (F12)
- Network tab shows all API calls
- See timing, headers, responses

## Common Issues

### "Cannot connect to runtime process"

**Problem:** Backend debugger can't start

**Solution:**
```bash
# Kill any running instances
pkill -f guardian-log
pkill -f air

# Try debug again
```

### "Chrome did not open"

**Problem:** Frontend debugger can't launch Chrome

**Solution:**
1. Check Chrome is installed
2. Or change to Edge in `launch.json`:
   ```json
   "type": "msedge"
   ```

### "Port already in use"

**Problem:** Server already running on port

**Solution:**
```bash
# Backend (8080)
kill $(lsof -ti:8080)

# Frontend (5173)
kill $(lsof -ti:5173)
```

### Breakpoints not hitting

**Problem:** Code execution doesn't stop

**Possible causes:**
1. Server not running in debug mode
2. Breakpoint in unreachable code
3. Source maps not working (frontend)

**Solution:**
1. Restart debug session
2. Check terminal for errors
3. Ensure code is actually being executed

## Advanced Debugging

### Remote Debugging (Delve)

For advanced Go debugging:

```bash
# Terminal 1: Start with Delve
dlv debug ./cmd/guardian-log --headless --listen=:2345 --api-version=2

# Terminal 2: Frontend
make dev-frontend

# VS Code: Use "Attach to Backend" configuration
# Or CLI: dlv connect :2345
```

### Debug Production Build

```bash
# Build with debug symbols
go build -o ./bin/guardian-log ./cmd/guardian-log

# Run
./bin/guardian-log

# Attach debugger
dlv attach $(pgrep guardian-log)
```

### Performance Profiling

```go
// Add to cmd/guardian-log/main.go
import _ "net/http/pprof"

// Add before server start:
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access profiles:
// http://localhost:6060/debug/pprof/
```

## VS Code Tasks

Run tasks from Command Palette (Ctrl+Shift+P) → "Tasks: Run Task"

Available tasks:
- **Start Backend (Hot Reload)** - Launch backend with Air
- **Start Frontend (Dev Server)** - Launch Vite dev server
- **Build Production** - Build production binary
- **Run Tests** - Execute test suite
- **Docker Build** - Build Docker image
- **Format Code** - Run formatters
- **Lint** - Run linters

## Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Start Debugging | F5 |
| Stop Debugging | Shift+F5 |
| Restart Debugging | Ctrl+Shift+F5 |
| Step Over | F10 |
| Step Into | F11 |
| Step Out | Shift+F11 |
| Continue | F5 |
| Toggle Breakpoint | F9 |
| Debug Console | Ctrl+Shift+Y |

## Stopping Debug Sessions

### Easy Cleanup

**Stop debugging:**
- Press **Shift+F5** in VS Code
- Or click red square in debug toolbar
- This stops the backend debugger

**Stop frontend dev server:**
- If it doesn't stop automatically, run task "Stop Frontend Dev Server"
- Or manually: `pkill -f 'vite'`

**Start fresh:**
- Servers might still be running in background
- Just press F5 again to restart debugging
- VS Code will reuse running servers

## Summary

**New simplified workflow:**

1. **Press F5** in VS Code
   - Select debug configuration
   - Servers start automatically

2. **Open browser manually**
   - Navigate to http://localhost:5173
   - No auto-launch

3. **Set breakpoints and debug**
   - Click line numbers
   - Use the app
   - Debugger stops at breakpoints

4. **Stop when done**
   - Press Shift+F5
   - Servers stop automatically

That's it! Maximum simplicity, full control.
