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
| SNMP Exporter | Synology NAS metrics | 9116 | Internal only |

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

# Install system-wide so the systemd unit can call it reliably.
sudo pip3 install podman-compose
podman-compose --version
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
sudo podman-compose up -d
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
sudo systemctl start observability-stack.service
sudo systemctl stop observability-stack.service
sudo systemctl restart observability-stack.service

# View status
sudo systemctl status observability-stack.service
sudo podman ps

# View logs
sudo podman logs grafana
sudo podman logs prometheus
sudo podman logs loki
sudo podman logs alloy
sudo podman logs caddy

# Restart individual service
cd /opt/observability
sudo podman-compose restart grafana
```

---

## Verification Checklist

```bash
# Check all containers running
sudo podman ps

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

## Troubleshooting

### Loki / Alloy Shows "unhealthy"

If `sudo podman ps` shows `loki` or `alloy` as `unhealthy`, it may be a healthcheck command issue rather than the service actually being down.

Recent `grafana/loki` and `grafana/alloy` images may not include `wget`/`curl`. If your `/opt/observability/compose.yaml` uses `wget` healthchecks, replace them with binary-based checks:

```yaml
loki:
  healthcheck:
    test: ["CMD", "loki", "-version"]

alloy:
  healthcheck:
    test: ["CMD", "alloy", "validate", "/etc/alloy/config.alloy"]
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
| Rootless containers | ⚠️ Not enabled (systemd unit runs Podman as root) |
| Systemd hardening | ✅ NoNewPrivileges, PrivateTmp, etc. |

---

## Landing Page

- **URL:** https://sentinella.alpina
- Beautiful dark-themed landing page with links to all services
- Shows system info, log sources, and status

---

## Grafana Dashboards

### Homelab Infrastructure Logs
- **UID:** homelab-logs
- Logs by Host pie chart
- Log volume over time
- Errors & Warnings panel
- Authentication events
- Pi-hole DNS logs
- All logs stream

### Homelab Overview
- **UID:** homelab-overview
- Summary of all systems
- Quick health indicators

### System Metrics
- **UID:** system-metrics
- CPU Usage (all hosts)
- Memory Usage (all hosts)
- Network I/O (rx/tx)
- Disk I/O (read/write)
- Disk Usage bar gauge
- Host Summary table

### NTP Time Synchronization
- **UID:** ntp-sync
- Time sync status
- Chrony source details

### OPNsense Firewall Security
- **UID:** opnsense-security
- Firewall events
- Blocked/allowed traffic

### Gotra Application
- **UID:** gotra-app
- Application logs
- Request patterns

### Proxmox VMs
- **UID:** proxmox-vms
- VM Overview table with status, CPU, Memory, Disk, Network
- CPU/Memory usage over time (per VM)
- Network and Disk I/O (per VM)
- Resource allocation summaries

### Alpina Homelab — Command Center
- **UID:** homelab-master
- Master dashboard with 70 panels across 10 sections
- Overview: Hosts Online, VMs Running, Avg CPU/Memory, NTP Sync, Log Entries
- Fleet Health: CPU/Memory/Disk bar gauges + time series for all hosts
- Proxmox Virtualization: VM status table
- NTP & Time Synchronization: Clock offset, sync status, frequency drift
- Pi-hole DNS & Ad Blocking: DNS query rate, CPU, memory, uptime
- OPNsense Firewall: Firewall events, TCP connections, log stream
- Network & Storage I/O: Network and disk I/O charts
- Logs & Events: Log volume, error rate, recent errors
- Home Assistant: Zigbee/Z-Wave device health
- **Portocali NAS (Xpenology):** System status, temps, volume usage stats (Vol2/Vol4), RAID/volume table, disk health, I/O, logs
- Host Inventory: Summary table with gauge columns

---

## Alerting Rules (Grafana)

Grafana unified alerting rules are provisioned for Portocali:
- Volume usage warning at 85% and critical at 90%
- RAID/pool degradation (any `synology_raid_status != 1`)

Rules are stored under a Grafana folder named `Portocali` and linked to the master dashboard panel `RAID / Volume Status` (panel id `207`).

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

### Portocali NAS (portocali.alpina) ✅
```bash
# syslog-ng configured (DSM 7.3.2 uses syslog-ng)
cat /etc/syslog-ng/patterndb.d/sentinella-remote.conf
# destination d_sentinella { udp("172.16.19.94" port(1514)); };
# log { source(src); destination(d_sentinella); };
```

---

## SNMP Exporter (Synology NAS Metrics)

Prometheus SNMP Exporter scrapes Synology-specific metrics from portocali NAS via SNMP.

| Setting | Value |
|---------|-------|
| Container | snmp-exporter (prom/snmp-exporter:v0.30.1) |
| Config | /opt/observability/snmp/snmp.yml |
| Module | synology |
| Auth | alpina_v2 (community: alpina) |
| Target | 172.16.21.21 (portocali) |
| Scrape interval | 120s |

Metrics collected include:
- synology_system_status, synology_temperature, synology_power_status
- synology_disk_status/temperature/health_status/bad_sector per disk (13 disks)
- synology_raid_status/free_size/total_size per RAID array (7 arrays)
- synology_service_users per service (CIFS, NFS, SSH, etc.)
- synology_storage_io_bytes_read/written per disk
- synology_space_io_bytes_read/written per volume

---

## PVE Exporter (Proxmox VM Metrics)

Prometheus PVE Exporter is installed on the Proxmox host for per-VM metrics.

| Setting | Value |
|---------|-------|
| Host | aria.alpina |
| Port | 9221 |
| Config | /etc/pve-exporter/pve.yml |
| API Token | root@pam!prometheus |

Metrics collected include:
- pve_cpu_usage_ratio - CPU utilization per VM
- pve_memory_usage_bytes / pve_memory_size_bytes - Memory usage
- pve_disk_size_bytes - Disk allocation
- pve_network_receive/transmit_bytes_total - Network I/O
- pve_disk_read/written_bytes_total - Disk I/O
- pve_uptime_seconds - VM uptime
- pve_guest_info - VM names and metadata

---

## Node Exporter (System Metrics)

All Linux hosts have Prometheus node_exporter installed for system metrics collection.

| Host | Port | Status |
|------|------|--------|
| sentinella.alpina | 9100 | ✅ Active |
| ntp.alpina | 9100 | ✅ Active |
| gotra | 9100 | ✅ Active |
| aria.alpina (Proxmox) | 9100 | ✅ Active |
| pihole | 9100 | ✅ Active |
| komga.alpina | 9100 | ✅ Active |

Metrics collected include:
- CPU usage per core and total
- Memory usage and availability
- Disk I/O and space usage
- Network I/O (bytes/packets)
- System load averages
- Uptime

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
sudo systemctl stop observability-stack.service

# Remove containers and volumes (WARNING: deletes data)
cd /opt/observability
sudo podman-compose down -v

# Remove config
sudo rm -rf /opt/observability
sudo rm /etc/systemd/system/observability-stack.service
sudo systemctl daemon-reload

# Remove DNS entries from Pi-hole
# Edit /etc/pihole/custom.list and remove sentinella entries
```
