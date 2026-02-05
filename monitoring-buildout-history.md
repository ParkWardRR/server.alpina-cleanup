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
