# Alpina Homelab Operations

Single source for network/host facts, access, IPv6 status, and app notes.

## Quick Links
- Firewall: OPNsense (gateway.alpina)
- DNS/Ad-block: Pi-hole (pihole.alpina)
- Monitoring: https://sentinella.alpina (Grafana/Prom/Loki/Alloy)
- Komga: http://komga.alpina (landing) / :25600 (UI)
- NTP: http://ntp.alpina (landing)
- Home Assistant: http://homeassistant.alpina:8123

## Monitoring Stack (Sentinella)
- Grafana: https://grafana.sentinella.alpina — user `admin`, pass `sG8pF8JcGVl4BypmiPy/j06HgMcPda41`
- Prometheus / Loki / Alloy (basic auth): https://prometheus.sentinella.alpina, https://loki.sentinella.alpina, https://alloy.sentinella.alpina — user `admin`, pass `vURLumGa0GMu4/nR2+vejcenAQBqt1un`
- Syslog ingest: `sentinella.alpina:1514/udp`
- Manage: `sudo systemctl restart observability-stack` (on sentinella); configs under `/opt/observability/`
- Portocali NAS:
  - Logs: `portocali.alpina` forwards syslog-ng to Alloy (`1514/udp`)
  - Metrics: Prometheus SNMP Exporter scrapes Synology/Xpenology OIDs (disk/RAID/temp/services/I/O)
  - Alerts (Grafana): volume usage >85%/>90% + RAID/pool degraded (folder `Portocali`)

## Hosts

| Host | IPv4 | IPv6 | Role |
|------|------|------|------|
| gateway.alpina | 172.16.16.16 | 2603:8001:7400:fa9a:a236:9fff:fe66:27ac | OPNsense router/RA/DHCPv6-PD |
| pihole | 172.16.66.66 | 2603:8001:7400:fa9a:4392:b645:21ad:5510 | DNS/Ad-block, DHCPv4 (DHCPv6 disabled) |
| ntp.alpina | 172.16.16.108 | 2603:8001:7400:fa9a:be24:11ff:fe60:2dfe | Chrony NTP, landing page |
| komga.alpina | 172.16.16.202 | 2603:8001:7400:fa9a:be24:11ff:fe09:c0b9 | Komga media server |
| home.alpina | 172.16.17.109 | 2603:8001:7400:fa9a:be24:11ff:fec9:2694 | General purpose server |
| sentinella.alpina | 172.16.19.94 | 2603:8001:7400:fa9a:be24:11ff:fe95:2956 | Observability stack |
| aria.alpina | 172.16.18.230 | 2603:8001:7400:fa9a:eaff:1eff:fed3:4683 | Proxmox host |
| homeassistant.alpina | 172.16.77.77 | — | HAOS appliance (no global IPv6 yet) |
| portocali.alpina | 172.16.21.21 | 2603:8001:7400:fa9a:7656:3cff:fe30:2dfc | NAS (Xpenology); SNMP metrics + syslog to Sentinella |

## SSH Access
```bash
ssh root@172.16.16.16                         # OPNsense
ssh pi@pihole                                  # Pi-hole
ssh alfa@ntp.alpina                            # NTP
ssh alfa@komga.alpina                          # Komga
ssh alfa@sentinella.alpina                     # Monitoring
ssh root@aria.alpina                           # Proxmox
ssh alfa@home.alpina                           # Home server
ssh alfa@portocali.alpina                      # NAS (Xpenology)
ssh root@homeassistant.local                   # Home Assistant (HAOS)
```

## IPv6 Status (as of 2026-02-06)
- Prefix: `2603:8001:7400:fa9a::/64` via DHCPv6-PD on OPNsense.
- OPNsense RAs: SLAAC + RDNSS (Pi-hole link-local) + DNSSL (alpina); firewall allows ICMPv6/NDP/DHCPv6.
- Pi-hole: dual-stack DNS; DHCPv6/RA **disabled** to stop rogue RAs.
- Working dual-stack hosts: gateway, pihole, ntp, komga, home, sentinella, aria (vmbr0), portocali — all with SLAAC.
- Outstanding: HAOS still lacks global IPv6; route-info-only RAs for ULA `fde6:19bd:3ffd::/64` from MACs `04:99:b9:71:36:9d`, `04:99:b9:84:18:17`, `ec:a9:07:07:22:bf`, `ac:bc:b5:db:26:fa` (identify/disable at source).
- Recent fixes (2026-02-06):
  - Proxmox: added `/etc/sysctl.d/50-ipv6.conf` with `accept_ra=2` (vmbr0/all); now has global address and internet v6 reachability.
  - home.alpina & sentinella.alpina: firewalld now allows `ipv6-icmp`; after clearing Pi-hole RA, single default via gateway and `ping -6` succeeds.
  - Pi-hole: `[dhcp] ipv6=false` in `pihole.toml`; restart FTL; no more default-route RAs from Pi-hole.
  - Portocali NAS: IPv6 enabled in DSM; SLAAC address `2603:...:fe30:2dfc` (EUI-64); default via gateway.
  - Sentinella: Observability stack (Caddy/Grafana/Prometheus/Loki/Alloy) now dual-stack; added `[::]:80/443/1514` port bindings and `enable_ipv6: true` on Podman network.

## Application Notes

### Komga (Debian 12, Docker)
- Status: Running Komga v1.24.1; NFS `/mnt/MonterosaSync-Read` mounted and persistent.
- Security: UFW enabled (SSH open; Komga 25600 limited to 172.16.0.0/16); SSH hardened; Fail2ban active; unattended-upgrades enabled.
- Access: http://komga.alpina (landing) and :25600 UI; SSH `alfa@komga.alpina`.
- Maintenance: reboot pending if new kernel installed; Komga auto-scans hourly; container restart policy `unless-stopped`.

### NTP (AlmaLinux 10)
- Chrony 4.6.1; 8 upstream sources; landing page at http://ntp.alpina; node_exporter on 9100.
- Backup: `/var/backups/ntp/pre-remediation-20260127_132804.tar.gz`.

### Pi-hole (v6)
- Upstreams: Quad9, Cloudflare, Google (IPv4+IPv6); `listeningMode=ALL`; `blockingmode=null`.
- Custom AAAA records added (core hosts) in `/etc/pihole/custom.list`; AAAA for `gateway.alpina` and service subdomains may be affected by known FTL bug.

### Backups
- OPNsense: `/root/ipv6-backup-20260205_213520/` (config, radvd, dhcp6c, rules, ndp snapshot).
- Pi-hole: `/home/pi/ipv6-backup-20260205_213526/` (pihole.toml, custom.list, sysctl, route snapshots).

### Open Actions
- Identify devices sending ULA route-info RAs; disable RA on those MACs.
- Add HAOS IPv6 (if possible) or document limitation.
- Add AAAA record for `portocali.alpina` in Pi-hole custom.list.
- Optional: request /56 PD from ISP (may change prefix; plan renumbering beforehand).
