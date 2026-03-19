# Getting Started with Onyx

## Install

### apt (Ubuntu / Debian — recommended for servers)

```bash
# Add the Onyx apt repository once
curl -fsSL https://elchi-dev.github.io/onyx/gpg.key \
  | sudo gpg --dearmor -o /usr/share/keyrings/onyx.gpg

echo "deb [signed-by=/usr/share/keyrings/onyx.gpg] \
  https://elchi-dev.github.io/onyx/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/onyx.list

sudo apt update && sudo apt install onyx
```

After this you'll get future updates with `sudo apt upgrade onyx`.

### curl (any Linux / macOS)

```bash
curl -fsSL https://elchi-dev.github.io/onyx/install.sh | bash
```

This detects your OS and architecture, downloads the right binary,
installs it to `/usr/local/bin`, optionally sets up a systemd service,
and runs `onyx setup` automatically.

### Build from source (Arch Linux / developers)

```bash
git clone https://github.com/Elchi-dev/onyx.git
cd onyx
go mod tidy
make build
sudo make install          # copies to /usr/local/bin/onyx
```

---

## First-time setup

If you built from source or skipped the installer's setup step:

```bash
onyx setup
```

The wizard asks for:
1. **Data directory** — where to store the database and config (`~/.config/onyx`)
2. **Dashboard password** — securely hashed, never stored in plain text
3. **HTTP port** — the port Onyx listens on (default: 80)
4. **Dashboard port** — the dashboard web UI port (default: 8080)
5. **First route** — optional, you can add more later

---

## Start

```bash
onyx start
```

Output:
```
→ routes loaded  config=0 database=1
→ Onyx starting  proxy=:80 dashboard=:8080
→ proxy listening addr=:80
→ dashboard listening addr=:8080
```

As a systemd service:
```bash
sudo systemctl start onyx
sudo systemctl status onyx
journalctl -u onyx -f
```

---

## Add routes

```bash
onyx route add --host api.example.com  --target http://localhost:3000
onyx route add --host app.example.com  --target http://localhost:5173
onyx route list
```

Or add them to `onyx.toml`:

```toml
[[routes]]
host    = "api.example.com"
target  = "http://localhost:3000"
enabled = true
```

---

## Open the dashboard

Visit `http://localhost:8080` and sign in with the password you set during setup.

You will see:
- **Live Traffic** — real-time WebSocket feed of every proxied request
- **Routes** — all registered routes and their status

---

## Test your proxy

Add a local DNS entry for testing:

```bash
echo "127.0.0.1  api.example.com" | sudo tee -a /etc/hosts
curl http://api.example.com
```

Watch the request appear in the dashboard live feed.
