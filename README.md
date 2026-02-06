# Alpina Homelab

Configuration and documentation for the Alpina homelab network.

## Quick Reference

| Service | Host | IP | Web Access |
|---------|------|-----|------------|
| Firewall | gateway.alpina | 172.16.16.16 | OPNsense Web UI |
| DNS/Ad-block | pihole | 172.16.66.66 | Pi-hole Admin |
| NTP | ntp.alpina | 172.16.16.108 | [Landing Page](http://ntp.alpina) |
| Home Assistant | homeassistant.alpina | 172.16.77.77 | [Web UI](http://homeassistant.alpina:8123) |
| Komga | komga.alpina | 172.16.16.202 | [Landing Page](http://komga.alpina) / [Komga UI](http://komga.alpina:25600) |
| Monitoring | sentinella.alpina | 172.16.19.94 | [Grafana](https://grafana.sentinella.alpina) |
| Proxmox | aria.alpina | 172.16.18.230 | Proxmox Web UI |

## Observability Stack

All infrastructure is monitored via the Sentinella observability stack:

- **Grafana Dashboards:** https://grafana.sentinella.alpina
  - [Alpina Homelab Command Center](https://grafana.sentinella.alpina/d/homelab-master) — Master dashboard with all key metrics
- **Prometheus Metrics:** 6 hosts with node_exporter + Proxmox PVE exporter (9 VMs)
- **Loki Logs:** Centralized syslog from Proxmox, Pi-hole, OPNsense, NTP, Gotra
- **Data Retention:** 24 months for both metrics and logs

## Server Landing Pages

| Server | URL | Features |
|--------|-----|----------|
| Komga | http://komga.alpina | System stats, 30-day CPU/memory charts, link to Komga UI |
| NTP | http://ntp.alpina | System stats, chrony sync status, NTP sources table |

## SSH Access

```bash
ssh root@172.16.16.16                         # OPNsense
ssh pi@pihole                                  # Pi-hole
ssh alfa@ntp.alpina                            # NTP server
ssh alfa@komga.alpina                          # Komga media server
ssh alfa@sentinella.alpina                     # Monitoring server
ssh root@aria.alpina                           # Proxmox host
ssh alfa@gotra                                 # Gotra app server
```

## Documentation

| Document | Description |
|----------|-------------|
| [homelab-details.md](homelab-details.md) | Network topology, service configs, custom DNS |
| [sentinella-observability-deployment.md](sentinella-observability-deployment.md) | Observability stack deployment and credentials |
| [monitoring-buildout-history.md](monitoring-buildout-history.md) | Monitoring build-out history and change log |
| [monitoring-roadmap.md](monitoring-roadmap.md) | Monitoring roadmap and future plans |
| [komga-remediation-plan.md](komga-remediation-plan.md) | Komga server hardening tasks |
| [ipv6-prep.md](ipv6-prep.md) | IPv6 preparation and rollout plan |
