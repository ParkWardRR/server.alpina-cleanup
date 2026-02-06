# Alpina Homelab

Configuration and documentation for the Alpina homelab network.

## Quick Reference

| Service | Host | IPv4 | IPv6 | Web Access |
|---------|------|------|------|------------|
| Firewall | gateway.alpina | 172.16.16.16 | 2603:8001:7400:fa9a:a236:9fff:fe66:27ac | OPNsense Web UI |
| DNS/Ad-block | pihole | 172.16.66.66 | 2603:8001:7400:fa9a:4392:b645:21ad:5510 | Pi-hole Admin |
| NTP | ntp.alpina | 172.16.16.108 | 2603:8001:7400:fa9a:be24:11ff:fe60:2dfe | [Landing Page](http://ntp.alpina) |
| Home Assistant | homeassistant.alpina | 172.16.77.77 | — | [Web UI](http://homeassistant.alpina:8123) |
| Komga | komga.alpina | 172.16.16.202 | 2603:8001:7400:fa9a:be24:11ff:fe09:c0b9 | [Landing Page](http://komga.alpina) / [Komga UI](http://komga.alpina:25600) |
| Monitoring | sentinella.alpina | 172.16.19.94 | 2603:8001:7400:fa9a:be24:11ff:fe95:2956 | [Grafana](https://grafana.sentinella.alpina) |
| Proxmox | aria.alpina | 172.16.18.230 | 2603:8001:7400:fa9a:eaff:1eff:fed3:4683 | Proxmox Web UI |
| Home Server | home.alpina | 172.16.17.109 | 2603:8001:7400:fa9a:be24:11ff:fec9:2694 | — |
| NAS | portocali.alpina | 172.16.21.21 | 2603:8001:7400:fa9a:7656:3cff:fe30:2dfc | DSM UI |

## Observability Stack

All infrastructure is monitored via the Sentinella observability stack:

- **Grafana Dashboards:** https://grafana.sentinella.alpina
  - [Alpina Homelab Command Center](https://grafana.sentinella.alpina/d/homelab-master) — Master dashboard with all key metrics
- **Prometheus Metrics:** 7 hosts with node_exporter + Proxmox PVE exporter (9 VMs)
- **NAS Metrics:** SNMP exporter for `portocali.alpina` (Synology/Xpenology OIDs)
- **Loki Logs:** Centralized syslog from Proxmox, Pi-hole, OPNsense, NTP, Gotra
- **NAS Logs:** syslog-ng from `portocali.alpina` to Alloy
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
ssh alfa@home.alpina                             # Home server
ssh alfa@portocali.alpina                        # NAS (Xpenology)
ssh root@homeassistant.local                     # Home Assistant (HAOS)
```

## Documentation

| Document | Description |
|----------|-------------|
| docs/ops.md | Single-source operations runbook (hosts, access, IPv6 snapshot, app notes) |
| docs/network-diagram.md | IPv6 network topology diagrams (rendered on GitHub) |
| docs/ipv6.md | IPv6 state, history, per-host status, remaining actions |
| ipv6-prep.md | Detailed IPv6 remediation plan & logs (kept for depth) |
| komga-remediation-plan.md | Komga server hardening log |
| README.md | You are here (quick reference) |
