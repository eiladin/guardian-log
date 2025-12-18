# VS Code Debug - Quick Reference

## ⚡ Quick Start

### 1. Just Press F5!
- Select **"Debug Backend"**, **"Debug Frontend"**, or **"Debug Full Stack"**
- Servers start automatically
- No manual terminal commands needed!

### 2. Set Breakpoints
- Click line number in any `.go` or `.tsx` file
- Red dot appears

### 3. Open Browser Manually
- Open http://localhost:5173 when you're ready
- Use the app to trigger your breakpoints

### 4. Debug!
- Debugger stops at breakpoints
- Inspect variables, step through code

## 🎯 Debug Options

| Configuration | What It Does | Browser URL |
|--------------|--------------|-------------|
| **Debug Backend** | Starts backend + frontend dev server | http://localhost:5173 |
| **Debug Frontend** | Starts frontend dev server only | http://localhost:5173 |
| **Debug Full Stack** | Starts backend + frontend dev server | http://localhost:5173 |

**Note:** All configurations start servers automatically. Just press F5 and open the browser when ready!

## 🔑 Keyboard Shortcuts

| Action | Key |
|--------|-----|
| Start Debug | F5 |
| Stop | Shift+F5 |
| Step Over | F10 |
| Step Into | F11 |
| Continue | F5 |
| Toggle Breakpoint | F9 |

## 💡 Pro Tips

1. **Just press F5** - servers start automatically!
2. **Open browser manually** - Navigate to http://localhost:5173 when ready
3. **Watch the terminal panel** - See server output and logs
4. **Use browser DevTools too** (F12) for network inspection
5. **Stop debugging** (Shift+F5) also stops the servers

## 📚 Full Guide

See [DEBUG_GUIDE.md](./DEBUG_GUIDE.md) for complete documentation.
