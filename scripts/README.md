# BotMatrix Deployment Scripts

This directory contains scripts for automating the deployment of the BotMatrix ecosystem to a remote server.

## 1. Quick Deployment (Recommended)

### Using Python (Cross-platform)
The `deploy.py` script is the most feature-rich and recommended method.

```bash
# Deploy EVERYTHING (Full Rebuild)
python scripts/deploy.py --target all

# Deploy only specific components (Fast Restart)
python scripts/deploy.py --target manager --fast
python scripts/deploy.py --target wxbot --fast
```

**Options:**
- `--ip <IP>`: Specify server IP (default: 192.168.0.167)
- `--user <USER>`: SSH username (default: derlin)
- `--target <TARGET>`: `all` (default), `manager`, `wxbot`
- `--fast`: Skip image rebuild, just restart (useful for code-only changes mounted in volumes, though usually rebuild is safer)

### Using PowerShell (Windows)
You can also use the PowerShell script directly:

```powershell
.\scripts\deploy.ps1
```

### Using Bash (Linux/Mac)
```bash
./scripts/deploy.sh
```

## 2. Debugging & Logs

### Watch Logs (Live)
Use the `watch_logs.py` script to view colored logs from the remote server.

```bash
python scripts/watch_logs.py
```
- Filters: `python scripts/watch_logs.py --filter "ERROR"`

### Windows Batch Shortcuts
For convenience, you can use the `.bat` files in the project root:
- `debug_logs.bat`: Tail logs via Docker
- `deploy_all.bat`: Run full deployment via PowerShell script

## 3. What the scripts do
1. **Bump Version**: Automatically increments patch version in `VERSION` file.
2. **Pack**: Zips all necessary source files (excluding venv, .git, etc.) using `pack_project.py`.
3. **Upload**: SCPs the zip to the remote server (`/tmp`).
4. **Deploy**:
   - SSH into the server.
   - Unzip to `/opt/BotMatrix`.
   - Run `docker-compose up -d --build`.
