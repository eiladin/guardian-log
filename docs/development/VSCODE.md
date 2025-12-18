# VS Code Setup for Guardian-Log

Complete Visual Studio Code configuration for Guardian-Log development with hot reload and debugging.

## Quick Start

### 1. Install Recommended Extensions

VS Code will prompt you to install recommended extensions from `.vscode/extensions.json`:

- **Go** (`golang.go`) - Go language support
- **ESLint** (`dbaeumer.vscode-eslint`) - JavaScript/TypeScript linting
- **Prettier** (`esbenp.prettier-vscode`) - Code formatting
- **JavaScript Debugger** (`ms-vscode.js-debug`) - Frontend debugging

### 2. Start Debugging

Just press **F5** and select a configuration!

**Available configurations:**
- **Debug Backend** - Starts both servers, debugs Go backend
- **Debug Frontend** - Starts frontend dev server only
- **Debug Full Stack** - Same as Debug Backend (best for API debugging)

## Debugging

For complete debugging documentation, see:
- **[Debug Guide](DEBUG_GUIDE.md)** - Complete debugging workflows and examples
- **[Quick Debug Reference](QUICK_DEBUG.md)** - One-page quick reference

### Quick Debug Steps

1. **Press F5** in VS Code
2. **Select configuration** (e.g., "Debug Backend")
3. **Wait for servers to start** (~5 seconds)
4. **Open browser** to http://localhost:5173
5. **Set breakpoints** by clicking line numbers
6. **Use the app** → breakpoints hit!

### Stop Debugging

- Press **Shift+F5**
- Or click red stop button
- Servers stop automatically

## Workspace Settings

Settings are configured in `.vscode/settings.json`:

**Go:**
- Auto-format on save with `gofmt`
- Run `go vet` on save
- Organize imports automatically

**TypeScript/React:**
- Prettier formatting on save
- ESLint integration
- Type checking

**Editor:**
- Format on save enabled
- Consistent indentation
- Optimized for project structure

## Tasks

Run tasks from Command Palette (`Ctrl+Shift+P` → "Tasks: Run Task"):

| Task | Description |
|------|-------------|
| Start Frontend Dev Server | Launch Vite dev server (background) |
| Stop Frontend Dev Server | Stop Vite dev server |
| Start Backend (Hot Reload) | Launch backend with Air |
| Build Production | Build production binary |
| Run Tests | Execute test suite |
| Docker Build | Build Docker image |
| Format Code | Run Go formatter |
| Lint | Run all linters |

## Hot Reload

Hot reload is built-in for development:

**Backend (Air):**
- Edit any `.go` file
- Save → Auto rebuild (~2-3 seconds)
- Backend restarts automatically

**Frontend (Vite):**
- Edit React components
- Save → Instant hot reload
- No page refresh needed!

Configuration files:
- Backend: `.air.toml`
- Frontend: `web/vite.config.ts`

## Terminal Integration

### Integrated Terminal

`Ctrl+` ` (backtick) to open terminal

**Recommended setup:**
- Terminal 1: Backend (`make dev-backend`)
- Terminal 2: Frontend (`make dev-frontend`)

### Task Terminal

Tasks run in dedicated terminals automatically.

## Keyboard Shortcuts

### Debugging
| Action | Shortcut |
|--------|----------|
| Start Debug | F5 |
| Stop Debug | Shift+F5 |
| Restart | Ctrl+Shift+F5 |
| Step Over | F10 |
| Step Into | F11 |
| Step Out | Shift+F11 |
| Continue | F5 |
| Toggle Breakpoint | F9 |

### Navigation
| Action | Shortcut |
|--------|----------|
| Go to File | Ctrl+P |
| Go to Symbol | Ctrl+Shift+O |
| Go to Definition | F12 |
| Find References | Shift+F12 |
| Command Palette | Ctrl+Shift+P |

### Editing
| Action | Shortcut |
|--------|----------|
| Format Document | Shift+Alt+F |
| Multi-cursor | Alt+Click |
| Rename Symbol | F2 |
| Quick Fix | Ctrl+. |

## File Watching

VS Code automatically watches for changes, but large projects may need adjustments.

**If you see "too many files" error:**

Linux:
```bash
echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**Excluded from watching** (configured in `.vscode/settings.json`):
- `node_modules/`
- `dist/`
- `tmp/`
- `bin/`
- `data/`

## Troubleshooting

### Debug Configuration Fails

**Issue:** "Cannot connect to runtime process"

**Solution:**
```bash
# Kill any running instances
pkill -f guardian-log
pkill -f air

# Try debug again (F5)
```

### Linter Not Working

**Issue:** Red squiggles not appearing

**Solution:**
1. Check extension is installed
2. Restart VS Code
3. Run `go mod download`

### Format on Save Not Working

**Issue:** Code not formatting when you save

**Solution:**
1. Check `"editor.formatOnSave": true` in settings
2. Verify formatter is installed (gofmt for Go, Prettier for TS)
3. Check output panel for errors

### Hot Reload Not Triggering

**Backend:**
```bash
# Check Air is running
ps aux | grep air

# Restart Air
pkill air
make dev-backend
```

**Frontend:**
```bash
# Check Vite is running
ps aux | grep vite

# Clear cache and restart
rm -rf web/node_modules/.vite
cd web && npm run dev
```

## Performance Tips

1. **Exclude large directories** from search:
   - Already configured in settings
   - Add more in `.vscode/settings.json` if needed

2. **Disable unused extensions**:
   - Disable extensions you don't use
   - Can slow down large projects

3. **Use workspace instead of folder**:
   - File → Open Workspace
   - Saves all settings and layout

4. **Close unused editors**:
   - `Ctrl+W` to close current tab
   - `Ctrl+K W` to close all

## Recommended Settings

Additional settings you may want to add:

```json
{
  // Auto-save
  "files.autoSave": "afterDelay",
  "files.autoSaveDelay": 1000,

  // Breadcrumbs
  "breadcrumbs.enabled": true,

  // Minimap
  "editor.minimap.enabled": false,

  // Terminal
  "terminal.integrated.fontSize": 14,

  // Git
  "git.autofetch": true,
  "git.confirmSync": false
}
```

Add these to `.vscode/settings.json` if desired.

## Extensions Setup

### Required Extensions

Install these for full functionality:

```bash
code --install-extension golang.go
code --install-extension dbaeumer.vscode-eslint
code --install-extension esbenp.prettier-vscode
```

### Optional But Recommended

```bash
code --install-extension eamodio.gitlens              # Git supercharged
code --install-extension github.copilot               # AI pair programmer
code --install-extension yzhang.markdown-all-in-one   # Markdown tools
```

## Summary

VS Code is fully configured for Guardian-Log development:

✅ **One-click debugging** - Press F5, servers start automatically
✅ **Hot reload** - Backend (2-3s) and frontend (instant)
✅ **Tasks** - Build, test, lint from Command Palette
✅ **Settings** - Optimized for Go and TypeScript
✅ **Extensions** - Recommended extensions auto-suggested

**Start developing:**
```bash
# Press F5 → Select "Debug Backend"
# Open http://localhost:5173
# Edit code → See changes live!
```

---

**[← Back to Development Guide](GUIDE.md)** • **[Debug Guide](DEBUG_GUIDE.md)** • **[Quick Debug Reference](QUICK_DEBUG.md)**
