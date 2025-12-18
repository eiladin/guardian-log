# Troubleshooting Guide

Common issues and solutions for Guardian-Log.

## Connection Issues

### Cannot Connect to AdGuard Home

**Symptoms:**
```
Failed to connect to AdGuard Home: authentication failed
```

**Solutions:**
1. Verify AdGuard Home is running
2. Check `AGH_URL` is correct
3. Verify credentials (`AGH_USER`, `AGH_PASS`)
4. Test connection: `curl -u $AGH_USER:$AGH_PASS $AGH_URL/control/status`

### Docker Network Issues

**For AdGuard on host machine:**

Linux:
```env
AGH_URL=http://172.17.0.1:8080
```

macOS/Windows:
```env
AGH_URL=http://host.docker.internal:8080
```

## API Issues

### Gemini API Errors

**Invalid API Key:**
```
Gemini API error: 401 Unauthorized
```

Solution: Get valid key at https://aistudio.google.com/app/apikey

**Rate Limit:**
```
Gemini API error: 429 Too Many Requests
```

Solution: Wait or upgrade to paid tier

## Database Issues

### Database Locked

```
database is locked
```

**Solution:**
```bash
docker-compose down
rm -f data/guardian.db-lock
docker-compose up -d
```

### Permission Errors

```
permission denied: data/guardian.db
```

**Solution:**
```bash
chmod 755 data/
chown 1000:1000 data/  # For Docker
```

## Build Issues

### Frontend Build Fails

**Error:** `npm ERR! code ENOENT`

**Solution:**
```bash
cd web
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Go Build Fails

**Error:** `go.mod requires go >= 1.25`

**Solution:**
```bash
# Install Go 1.25+
go version
```

## Docker Issues

### Port Already in Use

```
Error: bind: address already in use
```

**Solution:**
```bash
# Find process using port 8080
lsof -ti:8080

# Kill it
kill $(lsof -ti:8080)
```

### Container Won't Start

**Check logs:**
```bash
docker-compose logs guardian-log
```

**Common causes:**
- Missing `.env` file
- Invalid credentials
- Port conflict

## Performance Issues

### High Memory Usage

**Solution:**
- Increase `POLL_INTERVAL`
- Monitor with `docker stats`
- Check for memory leaks

### Slow Response

**Solution:**
- Check AdGuard Home response time
- Verify LLM API latency
- Reduce polling frequency

## No Anomalies Detected

**Possible reasons:**
1. All domains already in baseline
2. No new DNS queries
3. LLM disabled

**Solution:**
Reset baseline:
```bash
rm -f data/guardian.db
```

## Debug Mode

Enable debug logging:

```env
LOG_LEVEL=debug
```

View detailed logs:
```bash
docker-compose logs -f guardian-log
```

## Getting Help

If issues persist:

1. Check [GitHub Issues](https://github.com/OWNER/guardian-log/issues)
2. Enable debug logging
3. Collect logs
4. Open new issue with details

---

**[← Back to Docs](../README.md)**
