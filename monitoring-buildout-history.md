# Monitoring Stack Buildout History

## Overview
This document chronicles the deployment and configuration of the Sentinella observability platform for the Alpina homelab.

---

## Phase 1: Initial Stack Deployment (2026-02-05)

### Server Provisioning
- **Host:** sentinella.alpina (172.16.19.94)
- **OS:** AlmaLinux 10.1 (Heliotrope Lion)
- **Resources:** 4 vCPU, 8GB RAM, 70GB disk
- **Container Runtime:** Podman 5.6.0

### Components Deployed
| Component | Version | Purpose |
|-----------|---------|---------|
| Caddy | 2-alpine | Reverse proxy with internal TLS |
| Grafana | latest | Dashboards and visualization |
| Prometheus | latest | Metrics collection |
| Loki | latest | Log aggregation |
| Alloy | latest | Telemetry collector (syslog ingest) |

### Configuration Files Created
```
/opt/observability/
├── compose.yaml           # Podman compose definition
├── .env                   # Secrets (0600 permissions)
├── www/
│   └── index.html         # Landing page
├── caddy/
│   └── Caddyfile          # Reverse proxy config
├── prometheus/
│   └── prometheus.yml     # Scrape configuration
├── loki/
│   └── loki.yml           # Storage configuration
├── alloy/
│   └── config.alloy       # Syslog receiver config
└── grafana/
    └── provisioning/
        ├── datasources/
        │   └── datasources.yaml
        └── dashboards/
            ├── dashboards.yaml
            └── homelab-logs.json
```

### Technical Challenges Resolved
1. **Rootless Podman port binding** - Added `net.ipv4.ip_unprivileged_port_start=80` to sysctl
2. **SELinux blocking execution** - Moved podman-compose to /usr/local/bin for proper context
3. **Systemd user namespace issues** - Switched to root-level systemd service
4. **Syslog format mismatch** - Configured Alloy for RFC 3164 (BSD) format instead of RFC 5424

---

## Phase 2: DNS Configuration (2026-02-05)

### Pi-hole DNS Entries Added
Updated `/etc/pihole/pihole.toml` hosts array:
```
172.16.19.94 sentinella.alpina
172.16.19.94 grafana.sentinella.alpina
172.16.19.94 prometheus.sentinella.alpina
172.16.19.94 loki.sentinella.alpina
172.16.19.94 alloy.sentinella.alpina
```

---

## Phase 3: Log Source Integration (2026-02-05)

### Proxmox (aria.alpina)
- Installed rsyslog package
- Created `/etc/rsyslog.d/50-remote.conf`
- Forwarding all logs to sentinella.alpina:1514/udp
- Status: **Active**

### Pi-hole
- Installed rsyslog package
- Created `/etc/rsyslog.d/50-remote.conf`
- Forwarding all logs to sentinella.alpina:1514/udp
- Status: **Active**

### OPNsense (gateway)
- Documentation provided for web UI configuration
- Status: **Pending manual configuration**

---

## Phase 4: Landing Page & Dashboard (2026-02-05)

### Landing Page
- **URL:** https://sentinella.alpina/
- Dark-themed responsive design
- Links to all observability services
- System information display
- Log source status indicators

### Grafana Dashboard: Homelab Infrastructure Logs
- Pre-provisioned via file-based provisioning
- Panels:
  - Logs by Host (pie chart)
  - Log Volume Over Time (time series)
  - Errors & Warnings (log panel)
  - Authentication Events (log panel)
  - Pi-hole DNS Logs (log panel)
  - All Logs Stream (log panel)

---

## Phase 5: System Metrics with Node Exporter (2026-02-05)

### Node Exporter Deployment
Deployed Prometheus node_exporter to all Linux hosts for comprehensive system metrics:

| Host | Method | Port | Status |
|------|--------|------|--------|
| sentinella.alpina | Manual binary + systemd | 9100 | ✅ Active |
| ntp.alpina | Manual binary + systemd | 9100 | ✅ Active |
| gotra | Manual binary + systemd | 9100 | ✅ Active |
| aria.alpina (Proxmox) | apt package | 9100 | ✅ Active |
| pihole | apt package | 9100 | ✅ Active |

### Installation Steps
```bash
# For AlmaLinux hosts (manual binary)
wget https://github.com/prometheus/node_exporter/releases/download/v1.9.0/node_exporter-1.9.0.linux-amd64.tar.gz
tar xzf node_exporter-*.tar.gz
sudo mv node_exporter-*/node_exporter /usr/local/bin/
# Created systemd service at /etc/systemd/system/node_exporter.service

# For Debian hosts (apt package)
sudo apt install prometheus-node-exporter
```

### Firewall Configuration
Opened port 9100/tcp on all hosts to allow Prometheus scraping.

### Prometheus Configuration
Added node scrape job to `/opt/observability/prometheus/prometheus.yml`:
```yaml
- job_name: node
  static_configs:
    - targets:
        - sentinella.alpina:9100
        - ntp.alpina:9100
        - gotra:9100
        - aria.alpina:9100
        - pihole:9100
      labels:
        env: homelab
```

### System Metrics Dashboard
Created comprehensive Grafana dashboard with:
- CPU Usage over time (per host)
- Memory Usage over time (per host)
- Network I/O (rx/tx per host)
- Disk I/O (read/write per host)
- Disk Usage bar gauge
- Host Summary table with CPU%, Memory%, Disk%, Uptime, Load

---

## Phase 6: Proxmox VM Metrics (2026-02-05)

### PVE Exporter Deployment
Installed prometheus-pve-exporter on aria.alpina (Proxmox host) for per-VM metrics.

```bash
# Installation
pip3 install prometheus-pve-exporter --break-system-packages

# API Token created
pveum user token add root@pam prometheus --privsep=0

# Configuration at /etc/pve-exporter/pve.yml
# Systemd service at /etc/systemd/system/pve-exporter.service
# Listening on port 9221
```

### VMs Monitored
| VM ID | Name | Type |
|-------|------|------|
| 100 | Sentinella | QEMU |
| 101 | ezMaster | QEMU |
| 102 | Gotra | QEMU |
| 103 | Komga | QEMU |
| 105 | syncthings | QEMU |
| 106 | qacache | QEMU |
| 107 | Scriptwright | QEMU |
| 109 | Ottavia | QEMU |
| 110 | NTP | QEMU |

### Proxmox VMs Dashboard
Created comprehensive dashboard (uid: proxmox-vms) with:
- VM Overview table (Status, CPU%, Memory%, Disk, Uptime, Network)
- CPU Usage by VM (time series)
- Memory Usage by VM (time series)
- Memory Usage bar gauge (current)
- Disk Allocation bar gauge
- Network I/O by VM
- Disk I/O by VM
- Summary stats (Total Memory, Total Disk, Running VMs)

---

## Access Credentials

### Grafana
- URL: https://grafana.sentinella.alpina
- Username: admin
- Password: `sG8pF8JcGVl4BypmiPy/j06HgMcPda41`

### Prometheus / Loki / Alloy (Basic Auth)
- Username: admin
- Password: `vURLumGa0GMu4/nR2+vejcenAQBqt1un`

---

## Maintenance Commands

```bash
# SSH to server
ssh alfa@sentinella.alpina

# Service management
sudo systemctl start|stop|restart|status observability-stack

# View container status
sudo podman ps

# View container logs
sudo podman logs grafana|prometheus|loki|alloy|caddy

# Restart stack after config changes
sudo systemctl restart observability-stack
```

---

## Phase 7: Server Landing Pages (2026-02-05)

### Go-Based Landing Pages
Deployed custom Go landing pages with beautiful dark-themed designs, historical metrics from Prometheus, and service-specific information.

| Server | URL | Features |
|--------|-----|----------|
| NTP (ntp.alpina) | http://ntp.alpina:8080 | CPU/Memory charts, chrony sync status, NTP sources table |
| Komga (komga.alpina) | http://komga.alpina (port 80) | CPU/Memory charts, system stats, link to Komga UI |

### Technical Details
- Language: Go (compiled for Linux amd64)
- Charts: 30-day historical data via Prometheus API
- Design: Dark gradient theme with glassmorphism effects
- Service: systemd landing-page.service
- Port binding: setcap cap_net_bind_service for privileged ports

### NTP Landing Page Features
- Real-time chrony statistics (stratum, offset, drift)
- NTP source table with status indicators
- Historical CPU/memory usage from Prometheus

### Komga Landing Page Features
- Server resource utilization
- Direct link to Komga UI (port 25600)
- 30-day CPU and memory trends

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-02-05 | Initial stack deployment | Claude |
| 2026-02-05 | DNS configuration in Pi-hole v6 | Claude |
| 2026-02-05 | Proxmox rsyslog integration | Claude |
| 2026-02-05 | Pi-hole rsyslog integration | Claude |
| 2026-02-05 | Landing page creation | Claude |
| 2026-02-05 | Grafana dashboard provisioning | Claude |
| 2026-02-05 | OPNsense syslog integration | Claude |
| 2026-02-05 | NTP server rsyslog integration | Claude |
| 2026-02-05 | Gotra application rsyslog integration | Claude |
| 2026-02-05 | Added NTP, OPNsense, Gotra, Overview dashboards | Claude |
| 2026-02-05 | Increased retention to 24 months | Claude |
| 2026-02-05 | Landing page storage stats & predictions | Claude |
| 2026-02-05 | Deployed node_exporter to all Linux hosts | Claude |
| 2026-02-05 | Created System Metrics dashboard | Claude |
| 2026-02-05 | Enhanced Alloy with structured log parsing | Claude |
| 2026-02-05 | Deployed pve-exporter for Proxmox VM metrics | Claude |
| 2026-02-05 | Created Proxmox VMs dashboard with per-VM stats | Claude |
| 2026-02-05 | Deployed Go-based landing page on NTP server (port 8080) | Claude |
| 2026-02-05 | Deployed Go-based landing page on Komga server (port 80) | Claude |
| 2026-02-05 | Fixed komga.alpina node_exporter firewall (UFW port 9100) | Claude |
