# Debug Configuration Changes ✅

## What Changed

I've updated the VS Code debug configurations to be **much simpler**:

### Before (Manual)
1. Open Terminal 1 → `make dev-backend`
2. Open Terminal 2 → `make dev-frontend`
3. Wait for both to start
4. Press F5 in VS Code
5. Browser auto-launches

### After (Automatic)
1. **Press F5** in VS Code
2. **Open browser** when you're ready (http://localhost:5173)
3. That's it!

## What's Different

✅ **Servers start automatically** - No manual terminal commands
✅ **No browser auto-launch** - You control when to open the browser
✅ **Cleaner workflow** - Everything from VS Code
✅ **Same hot-reload** - Air and Vite still work perfectly

## How to Use

### Debug Backend (Go code)

```
1. Press F5
2. Select "Debug Backend"
3. Wait for servers to start (watch terminal panel)
4. Open http://localhost:5173 in browser
5. Set breakpoints in .go files
6. Use app → hits breakpoints!
```

### Debug Frontend (React code)

```
1. Press F5
2. Select "Debug Frontend"
3. Open http://localhost:5173
4. Press F12 for browser DevTools
5. Use Sources tab or console.log
```

### Debug Full Stack (Both)

```
1. Press F5
2. Select "Debug Full Stack"
3. Open http://localhost:5173
4. Set breakpoints in .go files
5. Use browser DevTools (F12) for React
6. Debug end-to-end!
```

## Debug Configurations Available

| Name | What It Does | Browser |
|------|-------------|---------|
| **Debug Backend** | Starts frontend + debugs Go backend | Manual: http://localhost:5173 |
| **Debug Frontend** | Starts frontend dev server | Manual: http://localhost:5173 |
| **Debug Full Stack** | Starts both + debugs Go backend | Manual: http://localhost:5173 |

## Files Changed

- ✅ `.vscode/launch.json` - Updated debug configs
- ✅ `.vscode/tasks.json` - Added auto-start tasks
- ✅ `.vscode/DEBUG_GUIDE.md` - Updated documentation
- ✅ `.vscode/QUICK_DEBUG.md` - Updated quick reference

## Try It Now!

```bash
# In VS Code:
1. Press F5
2. Select "Debug Backend"
3. Wait ~5 seconds for servers to start
4. Open http://localhost:5173 in your browser
5. Set a breakpoint in internal/api/handlers.go
6. Click something in the app
7. Breakpoint hits! 🎉
```

## Stopping

- **Shift+F5** to stop debugging
- Servers stop automatically
- Or run task: "Stop Frontend Dev Server"

## Need Help?

- **Quick Start:** [QUICK_DEBUG.md](./QUICK_DEBUG.md)
- **Full Guide:** [DEBUG_GUIDE.md](./DEBUG_GUIDE.md)
- **Main Docs:** [../DEVELOPMENT.md](../DEVELOPMENT.md)

---

**Bottom line:** Just press F5, wait a few seconds, open the browser, and debug! 🚀
