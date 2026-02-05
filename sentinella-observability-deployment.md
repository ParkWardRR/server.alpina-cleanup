# Sentinella Observability Stack Deployment

**Server:** sentinella.alpina (172.16.19.94)
**OS:** AlmaLinux 10.1 (Heliotrope Lion)
**Deployment Date:** 2026-02-05
**Container Runtime:** Podman 5.6.0 + podman-compose 1.5.0
**Status:** ✅ COMPLETE

---

## Stack Components

| Service | Purpose | Internal Port | External Access |
|---------|---------|---------------|-----------------|
| Caddy | Reverse proxy + TLS | 80, 443 | https://*.alpina |
| Grafana | Dashboards & visualization | 3000 | https://grafana.sentinella.alpina |
| Prometheus | Metrics collection | 9090 | https://prometheus.sentinella.alpina |
| Loki | Log aggregation | 3100 | https://loki.sentinella.alpina |
| Alloy | Telemetry collector | 12345, 1514/udp | https://alloy.sentinella.alpina |

---

## Access URLs & Credentials

### Grafana (Main Dashboard)
- **URL:** https://grafana.sentinella.alpina
- **Username:** admin
- **Password:** `sG8pF8JcGVl4BypmiPy/j06HgMcPda41`

### Prometheus, Loki, Alloy (Basic Auth)
- **URLs:**
  - https://prometheus.sentinella.alpina
  - https://loki.sentinella.alpina
  - https://alloy.sentinella.alpina
- **Username:** admin
- **Password:** `vURLumGa0GMu4/nR2+vejcenAQBqt1un`

### Syslog Ingest
- **Endpoint:** sentinella.alpina:1514/udp
- **Usage:** Configure devices to send syslog to this address

---

## Deployment Log

### Phase 1: System Preparation
```bash
# Verified SSH and sudo access
ssh alfa@sentinella.alpina
sudo whoami  # root

# System specs: 8GB RAM, 70GB disk, 4 vCPU
# Podman 5.6.0 already installed
```

### Phase 2: Install podman-compose
```bash
sudo dnf install -y python3-pip
pip3 install --user podman-compose
# Installed podman-compose 1.5.0
```

### Phase 3: Create Directory Structure
```bash
sudo mkdir -p /opt/observability/{caddy,prometheus,loki,alloy}
sudo chown -R alfa:alfa /opt/observability
```

### Phase 4: Generate Secrets
```bash
# Created /opt/observability/.env with:
# - Grafana admin password (random 32 char)
# - Basic auth credentials (random 32 char)
# - bcrypt hash for Caddy basic auth
chmod 600 /opt/observability/.env
```

### Phase 5: Create Configuration Files

**Files created:**
- `/opt/observability/compose.yaml` - Podman compose file
- `/opt/observability/caddy/Caddyfile` - Reverse proxy config with TLS
- `/opt/observability/prometheus/prometheus.yml` - Metrics scraping config
- `/opt/observability/loki/loki.yml` - Log storage config
- `/opt/observability/alloy/config.alloy` - Syslog receiver config

### Phase 6: Enable Unprivileged Ports
```bash
echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
sudo sysctl --system
```

### Phase 7: Configure Firewall
```bash
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-port=1514/udp
sudo firewall-cmd --reload
```

### Phase 8: Create Systemd Service
```bash
# Created /etc/systemd/system/observability-stack.service
sudo systemctl daemon-reload
sudo systemctl enable observability-stack.service
```

### Phase 9: Start Stack
```bash
cd /opt/observability
~/.local/bin/podman-compose up -d
```

### Phase 10: Add DNS Entries
```bash
# Added to Pi-hole custom.list:
172.16.19.94 grafana.sentinella.alpina
172.16.19.94 prometheus.sentinella.alpina
172.16.19.94 loki.sentinella.alpina
172.16.19.94 alloy.sentinella.alpina
172.16.19.94 sentinella.alpina
```

---

## Architecture

```
                    Internet/LAN
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │  Firewall (firewalld)                  │
    │  Ports: 80/tcp, 443/tcp, 1514/udp      │
    └────────────────────────────────────────┘
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │  Caddy (Reverse Proxy)                 │
    │  - TLS termination (internal CA)       │
    │  - Basic auth for Prom/Loki/Alloy      │
    │  - Routes *.alpina to services         │
    └────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    ┌─────────┐    ┌──────────┐    ┌─────────┐
    │ Grafana │    │Prometheus│    │  Loki   │
    │  :3000  │    │  :9090   │    │ :3100   │
    └─────────┘    └──────────┘    └─────────┘
         │               │               ▲
         │               │               │
         └───────────────┼───────────────┘
                         │
                    ┌────┴────┐
                    │  Alloy  │
                    │ :12345  │◄──── Syslog :1514/udp
                    └─────────┘
```

---

## File Locations

| File | Purpose |
|------|---------|
| `/opt/observability/compose.yaml` | Podman compose definition |
| `/opt/observability/.env` | Secrets (0600 permissions) |
| `/opt/observability/caddy/Caddyfile` | Reverse proxy config |
| `/opt/observability/prometheus/prometheus.yml` | Prometheus scrape config |
| `/opt/observability/loki/loki.yml` | Loki storage config |
| `/opt/observability/alloy/config.alloy` | Alloy syslog config |
| `/etc/systemd/system/observability-stack.service` | Systemd unit |

---

## Management Commands

```bash
# SSH to server
ssh alfa@sentinella.alpina

# Start/stop/restart stack
sudo systemctl start observability-stack
sudo systemctl stop observability-stack
sudo systemctl restart observability-stack

# View status
sudo systemctl status observability-stack
podman ps

# View logs
podman logs grafana
podman logs prometheus
podman logs loki
podman logs alloy
podman logs caddy

# Restart individual service
cd /opt/observability
~/.local/bin/podman-compose restart grafana
```

---

## Verification Checklist

```bash
# Check all containers running
podman ps

# Test Grafana
curl -k https://grafana.sentinella.alpina/api/health

# Test Prometheus
curl -k -u admin:PASSWORD https://prometheus.sentinella.alpina/-/healthy

# Test Loki
curl -k -u admin:PASSWORD https://loki.sentinella.alpina/ready

# Test Alloy
curl -k -u admin:PASSWORD https://alloy.sentinella.alpina/-/ready

# Test syslog ingest
echo "<14>Test message from CLI" | nc -u sentinella.alpina 1514
```

---

## Grafana Data Sources (Configure in UI)

### Prometheus
- Type: Prometheus
- URL: http://prometheus:9090
- Access: Server (default)

### Loki
- Type: Loki
- URL: http://loki:3100
- Access: Server (default)

---

## Security Features

| Feature | Status |
|---------|--------|
| TLS encryption | ✅ Internal CA (self-signed) |
| Basic auth on Prometheus/Loki/Alloy | ✅ Enabled |
| Grafana authentication | ✅ Built-in |
| Firewall | ✅ Only 80, 443, 1514/udp exposed |
| Internal ports not exposed | ✅ 3000, 9090, 3100, 12345 internal only |
| Secrets file permissions | ✅ 0600 |
| Rootless containers | ✅ Running as user alfa |
| Systemd hardening | ✅ NoNewPrivileges, PrivateTmp, etc. |

---

## Landing Page

- **URL:** https://sentinella.alpina
- Beautiful dark-themed landing page with links to all services
- Shows system info, log sources, and status

---

## Grafana Dashboard

**Dashboard:** Homelab Infrastructure Logs
- Pre-provisioned via `/opt/observability/grafana/provisioning/`
- Includes:
  - Logs by Host pie chart
  - Log volume over time
  - Errors & Warnings panel
  - Authentication events
  - Pi-hole DNS logs
  - All logs stream

---

## Configured Log Sources

### Proxmox (aria.alpina) ✅
```bash
# rsyslog installed and configured
cat /etc/rsyslog.d/50-remote.conf
# *.* @sentinella.alpina:1514
```

### Pi-hole ✅
```bash
# rsyslog installed and configured
cat /etc/rsyslog.d/50-remote.conf
# *.* @sentinella.alpina:1514
```

### OPNsense (gateway) ✅
```bash
# Configured via /etc/syslog.d/sentinella.conf
ssh root@172.16.16.16
cat /etc/syslog.d/sentinella.conf
# *.*    @172.16.19.94:1514
```

### NTP Server (ntp.alpina) ✅
```bash
# rsyslog configured
cat /etc/rsyslog.d/50-remote.conf
# *.* @172.16.19.94:1514
```

### Gotra Application ✅
```bash
# rsyslog configured for system + container logs
cat /etc/rsyslog.d/50-remote.conf
# *.* @172.16.19.94:1514
```

---

## Sending Logs from Other Devices

### From Linux (rsyslog)
```bash
echo '*.* @sentinella.alpina:1514' | sudo tee /etc/rsyslog.d/50-remote.conf
sudo systemctl restart rsyslog
```

### From Network Devices
Configure syslog destination: `172.16.19.94:1514 UDP`

---

## Maintenance Notes

- Prometheus retains data for 24 months (730 days)
- Loki retains data for 24 months (17520 hours)
- Containers restart automatically on failure
- Stack starts on boot via systemd
- TLS certificates are self-signed (browser warning expected)

---

## Rollback Procedure

```bash
# Stop stack
sudo systemctl stop observability-stack

# Remove containers and volumes (WARNING: deletes data)
cd /opt/observability
~/.local/bin/podman-compose down -v

# Remove config
sudo rm -rf /opt/observability
sudo rm /etc/systemd/system/observability-stack.service
sudo systemctl daemon-reload

# Remove DNS entries from Pi-hole
# Edit /etc/pihole/custom.list and remove sentinella entries
```
