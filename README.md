# Alpina Homelab - Network Infrastructure

Configuration management and documentation for the Alpina homelab network infrastructure.

## Network Overview

| Device | Hostname | IP Address | Role |
|--------|----------|------------|------|
| OPNsense | gateway.alpina | 172.16.16.16 | Firewall/Router |
| Pi-hole | pihole | 172.16.66.66 | DNS/Ad-blocking |
| NTP Server | ntp.alpina | 172.16.16.108 | Time synchronization |

## Network Topology

```
Internet
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│  OPNsense Firewall (gateway.alpina)                        │
│  WAN: 172.88.97.187 (DHCP from ISP)                        │
│  LAN: 172.16.16.16/16                                       │
│  IPv6: 2603:8001:7402:cf1c::/64                            │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│  LAN (172.16.0.0/16)                                        │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Pi-hole    │  │ NTP Server  │  │   Clients   │         │
│  │ .66.66      │  │ .16.108     │  │   Various   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Quick Access (SSH)

```bash
# OPNsense firewall
ssh root@172.16.16.16

# Pi-hole DNS server
ssh pi@pihole

# NTP time server
ssh alfa@ntp.alpina
```

## Services

### OPNsense (gateway.alpina)
- **Version:** 23.7.12_5 (FreeBSD 13.2) ⚠️ *Update to 25.1.x recommended*
- **Hardware:** AMD AM1 platform (Athlon/Sempron Kabini APU), 8GB DDR3, 100GB SSD
  - NICs: 4x Intel I350/I210 GbE (PCIe) + 1x Realtek (onboard, unused)
- **Interfaces:**
  - LAN (igb0): 172.16.16.16/16
  - WAN (igb3): DHCP from Charter/Spectrum
- **Services:** Firewall (pf), NAT, Suricata IDS, DHCP Relay, NTP
- **IDS:** Suricata with Emerging Threats (ET/Open) and abuse.ch rules
  - Rule Sources: ET/Open (~20k+ rules), Feodo Tracker, SSL Blacklist, URLhaus, ThreatFox
  - Mode: IDS (passive monitoring)
  - Interfaces: LAN + WAN (optimized 2026-01-27)
  - Pattern Matcher: ac-ks (optimized for AMD AM1)
  - Updates: Daily automatic

### Pi-hole (pihole)
- **Version:** v6.3 (Core), v6.4 (Web), v6.4.1 (FTL)
- **Hardware:** Raspberry Pi 3 Model B (ARM Cortex-A53, 1GB RAM, 32GB microSD)
- **OS:** Raspberry Pi OS (Debian Bookworm, aarch64)
- **IP:** 172.16.66.66
- **Services:** DNS (port 53), DHCP (port 67), Web UI (80/443)
- **Upstream DNS:** Quad9 (9.9.9.11), Cloudflare (1.1.1.1)

### NTP Server (ntp.alpina)
- **Hardware:** Proxmox VE virtual machine
- **OS:** Enterprise Linux 10 (Rocky/Alma Linux)
- **IP:** 172.16.16.108
- **Service:** Chrony 4.6.1 (Stratum 2, synced to time2.google.com)
- **Upstream:** Google, NIST, Cloudflare time servers

## Documentation

- [Homelab Inventory](docs/homelab-inventory.md) - Detailed hardware/software inventory
- [History](docs/history.md) - Change log and incident history
- [Action Plan](docs/action-plan.md) - Pending fixes and improvements

## Repository Structure

```
.
├── README.md                 # This file
├── spec.md                   # Original analysis request
├── docs/
│   ├── homelab-inventory.md  # Detailed inventory + Suricata IDS config
│   ├── history.md            # Change history
│   └── action-plan.md        # Pending actions + IDS optimization tasks
└── .context/
    └── notes.md              # Context notes
```

## IDS Status Summary

| Component | Status |
|-----------|--------|
| Suricata Service | ✅ Running (2.9GB RAM) |
| Emerging Threats Rules | ✅ Active, daily updates |
| abuse.ch Feeds | ✅ Active (Feodo, SSL, URLhaus, ThreatFox) |
| WAN Monitoring | ✅ Enabled (2026-01-27) |
| Pattern Matcher | ✅ ac-ks (optimized for AMD AM1) |
| Recent Alerts | ✅ None (clean traffic) |

See [docs/homelab-inventory.md](docs/homelab-inventory.md#suricata-ids-configuration) for full IDS configuration details and optimization recommendations.
