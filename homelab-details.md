# Alpina Homelab - Technical Details

## Network Overview

| Device | Hostname | IP Address | Role |
|--------|----------|------------|------|
| OPNsense | gateway.alpina | 172.16.16.16 | Firewall/Router |
| Pi-hole | pihole | 172.16.66.66 | DNS/Ad-blocking |
| NTP Server | ntp.alpina | 172.16.16.108 | Time synchronization |
| Komga | komga.alpina | 172.16.16.202 | Media server (comics) |

## Network Topology

```
Internet
    |
    v
+-----------------------------------------------------------+
|  OPNsense Firewall (gateway.alpina)                       |
|  WAN: 172.88.97.187 (DHCP from ISP)                       |
|  LAN: 172.16.16.16/16                                     |
|  IPv6: 2603:8001:7402:cf1c::/64                           |
+-----------------------------------------------------------+
    |
    v
+-----------------------------------------------------------+
|  LAN (172.16.0.0/16)                                      |
|                                                           |
|  +-------------+  +-------------+  +-------------+        |
|  |  Pi-hole    |  | NTP Server  |  |   Komga     |        |
|  | .66.66      |  | .16.108     |  | .16.202     |        |
|  +-------------+  +-------------+  +-------------+        |
+-----------------------------------------------------------+
```

## SSH Access

```bash
# OPNsense firewall
ssh root@172.16.16.16

# Pi-hole DNS server
ssh -i ~/.ssh/id_ed25519 pi@pihole

# NTP time server
ssh -i ~/.ssh/id_ed25519 alfa@ntp.alpina

# Komga media server
ssh -i ~/.ssh/id_ed25519_komga_alpina alfa@komga.alpina
```

---

## OPNsense (gateway.alpina)

**Version:** 23.7.12_5 (FreeBSD 13.2)
**Hardware:** AMD AM1 platform, 8GB DDR3, 100GB SSD

### Network Interfaces
- LAN (igb0): 172.16.16.16/16
- WAN (igb3): DHCP from Charter/Spectrum

### Services
- Firewall (pf), NAT, DHCP Relay, NTP
- Suricata IDS with Emerging Threats and abuse.ch rules

---

## Pi-hole (pihole)

**Version:** Core v6.3, Web v6.4, FTL v6.4.1
**Hardware:** Raspberry Pi 3 Model B (ARM Cortex-A53, 1GB RAM)
**OS:** Raspberry Pi OS (Debian Bookworm, aarch64)

### DNS Configuration
- Upstream DNS: Quad9 (9.9.9.11), Cloudflare (1.1.1.1), Google (8.8.8.8)
- Blocklists: StevenBlack hosts (~72k domains)
- ESNI blocking, Mozilla canary, iCloud Private Relay blocking enabled

### Custom DNS Records
```
172.16.66.66    pihole.alpina
172.16.20.30    zagato.alpina
172.16.21.21    portocali
172.16.21.69    munich
172.16.21.185   my.ampache.alpina
172.16.20.20    trofeo
172.16.20.20    engenius
172.16.19.77    meshcentral.alpina
172.16.18.230   aria
172.16.18.230   aria.alpina
172.16.20.57    ezmaster
172.16.16.108   ntp
```

### Pending Tasks
- Reboot for kernel update (6.12.62)
- Consider enabling DNSSEC
- Harden SSH (disable password auth)
- Implement IPv6

---

## NTP Server (ntp.alpina)

**Hardware:** Proxmox VE VM (Intel N100, 1.7GB RAM)
**OS:** AlmaLinux 10.1
**Service:** Chrony 4.6.1 (Stratum 2)

### Time Sources (8 configured)
- NIST: time-d-wwv.nist.gov (Stratum 1)
- Google: time.google.com, time1-3.google.com (Stratum 1)
- Cloudflare: time.cloudflare.com
- Pool: 2x pool.ntp.org

### Hardening Applied
- Firewall: NTP allowed, Cockpit removed
- Auto-updates: dnf-automatic enabled
- Health monitoring: 5-minute timer
- Logging: /var/log/chrony with rotation
- Rate limiting: DDoS protection enabled

### Backup Location
```
/var/backups/ntp/pre-remediation-20260127_132804.tar.gz
```

---

## Komga (komga.alpina)

**Hardware:** Proxmox VE VM (4 vCPU Intel N100, 4.2GB RAM, 392GB disk)
**OS:** Debian 12 (bookworm)
**Service:** Komga 1.19.1 (Docker)

### Configuration
- Port: 25600 (web UI)
- Data: /mnt/MonterosaSync-Read (NFS mount)
- Config: /home/alfa/komga/config
- Library: 78 series, 660 books

### Status
See [komga-remediation-plan.md](komga-remediation-plan.md) for pending hardening tasks.
