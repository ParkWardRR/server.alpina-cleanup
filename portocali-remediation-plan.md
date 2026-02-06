# Portocali NAS Remediation Plan

**Server:** portocali.alpina (172.16.21.21)
**Platform:** Xpenology DS3622xs+ (Arc bootloader beta) / DSM 7.3.2-86009
**Hardware:** Intel i7-13700T (24 threads), 125GB RAM, 11 SATA + 2 NVMe
**Assessment Date:** 2026-02-05
**Remediation Date:** 2026-02-05
**Status:** ✅ COMPLETE (follow-ups noted)

---

## Executive Summary

| Category | Status | Priority |
|----------|--------|----------|
| SSH Hardening | ✅ Complete | HIGH |
| NTP Sync | ✅ Using ntp.alpina | MEDIUM |
| SNMP Metrics (Prometheus) | ✅ Active via snmp-exporter | HIGH |
| Log Forwarding (Loki) | ✅ syslog-ng → Sentinella | HIGH |
| Grafana Dashboard | ✅ 13-panel NAS section added | HIGH |
| Storage Capacity | ⚠️ volumes 2 & 4 at 86-88% | CRITICAL |
| Auto-Updates | ⚠️ N/A (Xpenology — NEVER auto-update) | INFO |

---

## System Profile

| Component | Value |
|-----------|-------|
| Hostname | Portocali |
| IP | 172.16.21.21 |
| Platform | Xpenology DS3622xs+ (Arc bootloader) |
| DSM | 7.3.2-86009 (GM, 2025-11-26) |
| Kernel | 4.4.302+ |
| CPU | Intel i7-13700T (24 threads) |
| RAM | 125GB DDR |
| Drives | 11 SATA (sda-sdk) + 2 NVMe |
| NIC | Intel e1000e, 172.16.21.21/16 |

### Storage Layout

| Array | Type | Size | Used | Mount | Status |
|-------|------|------|------|-------|--------|
| md0 | RAID1 (3/17) | 8GB | 19% | / (system) | ✅ Healthy |
| md2 | RAID10 (4/4) | 7.0TB | 42% | /volume1 | ✅ OK |
| md4 | RAID1 (NVMe) | 885GB | <1% | /volume3 | ✅ OK |
| md6 | RAID5 (3/3) | 32TB | 86% | /volume2 | ⚠️ Warning |
| md5 | RAID5 (4/4) | 37TB | 88% | /volume4 | ⚠️ Warning |

### Disk Inventory (from SNMP)

| Disk | Model | Notes |
|------|-------|-------|
| Disk 1 | HUS724040ALE641 | 93 bad sectors — monitor closely |
| Disk 2 | OOS4000G | |
| Disk 3 | HUS724040ALE641 | |
| Disk 4 | ST18000NM003D-3DL103 | 18TB |
| Disk 5 | WD140EDFZ-11A0VA0 | 14TB |
| Disk 6 | ST18000NM003D-3DL103 | 18TB |
| Disk 7 | ST14000NM005G-2KG133 | 14TB |
| Disk 8 | WD140EDFZ-11A0VA0 | 14TB |
| Disk 9 | HUS724040ALE641 | |
| Disk 10 | ST18000NM003D-3DL103 | 18TB |
| Disk 11 | WD140EDFZ-11A0VA0 | 14TB |
| M.2 Drive 1-1 | SAMSUNG MZVPV256HDGL | NVMe 256GB |
| M.2 Drive 2-1 | WD Green SN350 1TB | NVMe 1TB |

### Running Services

| Service | Port | Active Users |
|---------|------|-------------|
| CIFS/SMB | 139, 445 | 1 |
| NFS | 2049 | 5 |
| SSH | 22 | 1 |
| HTTP/HTTPS | 80, 443, 5000, 5001 | 1 |
| Plex | 32400 | — |
| iSCSI | 3261-3265 | — |
| Syncthing | 5566 | — |
| SNMP | 161 (LAN) | Prometheus |

---

## Execution Summary

Core remediation completed on 2026-02-05 (follow-ups below):

### 1. SSH Hardening ✅

Changes to `/etc/ssh/sshd_config`:
```
PermitRootLogin no
MaxAuthTries 3
PubkeyAuthentication yes
X11Forwarding no
PasswordAuthentication yes  (left enabled — see note below)
```

**Note:** Password auth deliberately left enabled. Locking out of a NAS is catastrophic. Disable after confirming key auth is reliable:
```bash
sudo sed -i 's/^PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo synosystemctl restart sshd  # or kill -HUP $(pidof sshd)
```

### 2. NTP Configuration ✅

Updated `/etc/synoinfo.conf`:
```
ntpdate_server="172.16.16.108"        # ntp.alpina (Stratum 2)
ntpdate_server_backup="time.google.com"
```

Verified sync: `adjust time server 172.16.16.108 offset +0.001404 sec`

### 3. SNMP Configuration ✅

Updated `/etc/snmp/snmpd.conf`:
```
rocommunity alpina 172.16.0.0/16
rocommunity alpina 127.0.0.1
agentAddress udp:0.0.0.0:161
sysLocation  Alpina Homelab
sysContact   alfa@portocali.alpina
sysName      Portocali
```

Boot persistence script at `/usr/local/bin/start-snmpd.sh`.

### 4. Syslog Forwarding ✅

Created `/etc/syslog-ng/patterndb.d/sentinella-remote.conf`:
```
destination d_sentinella {
    udp("172.16.19.94" port(1514));
};
log {
    source(src);
    destination(d_sentinella);
};
```

Verified: `{host="Portocali"}` visible in Loki alongside aria, gateway, gotra, ntp, pihole.

### 5. SNMP Exporter on Sentinella ✅

Added `snmp-exporter` container to `/opt/observability/compose.yaml`:
```yaml
snmp-exporter:
  image: docker.io/prom/snmp-exporter:v0.30.1
  container_name: snmp-exporter
  restart: unless-stopped
  volumes:
    - ./snmp/snmp.yml:/etc/snmp_exporter/snmp.yml:ro
  networks:
    - observability
```

Custom `snmp.yml` at `/opt/observability/snmp/snmp.yml` with Synology module:
- synoSystem (temp, power, fans, model, DSM version)
- synoDisk (health, temp, bad sectors, remaining life per disk)
- synoRaid (status, free/total size per array)
- synoUPS (battery, load, runtime)
- synologyService (active users per service)
- storageIO (bytes read/written, load per disk)
- spaceIO (bytes read/written, load per volume)

### 6. Prometheus Scrape Config ✅

Added to `/opt/observability/prometheus/prometheus.yml`:
```yaml
- job_name: portocali-snmp
  scrape_interval: 120s
  scrape_timeout: 60s
  metrics_path: /snmp
  params:
    auth: [alpina_v2]
    module: [synology]
  static_configs:
    - targets: ['172.16.21.21']
      labels:
        hostname: portocali.alpina
        role: nas
        os: xpenology-dsm
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - source_labels: [__param_target]
      target_label: instance
    - target_label: __address__
      replacement: snmp-exporter:9116
```

Verified: target is UP, scrape duration ~1.05s, 471 PDUs returned.

### 7. Master Dashboard Updated ✅

Added "Portocali NAS (Xpenology)" section to homelab-master dashboard with 13 panels:

| Panel | Type | What it shows |
|-------|------|--------------|
| System Status | stat | 1=Normal, 2=Failed |
| System Temp | stat | Celsius with color thresholds |
| Power Status | stat | 1=Normal, 2=Failed |
| Fan Status | stat | System fan health |
| Active Users | stat | SMB + NFS user counts |
| Total Disks | stat | Disk count |
| RAID / Volume Status | table | Per-volume status, free/total/used% with gauge |
| Disk Health | table | Per-disk health, temp, bad sectors with color coding |
| Disk I/O | timeseries | Per-disk read/write bytes/sec |
| Volume I/O | timeseries | Per-volume read/write bytes/sec |
| Temperatures | timeseries | System + per-disk temperature history |
| NAS Logs | logs | Live syslog stream from Loki |

---

## Xpenology / Arc Bootloader Considerations

**CRITICAL:** This is NOT a genuine Synology. It runs Xpenology with the Arc bootloader beta.

- **NEVER run DSM Update** — DSM updates will break the Arc bootloader
- **Package Center updates are safe** — individual packages can be updated
- System-level config changes may be overwritten by DSM internal processes
- Use Synology Task Scheduler for persistent scheduled tasks (not raw crontab)
- No systemd — services are managed by DSM's init scripts
- Kernel is 4.4.x (old) — some modern tools may not work

---

## DSM Boot Tasks (Persistence)

**Create in DSM: Control Panel → Task Scheduler → Create → Triggered Task → User-defined script**

| Task Name | Event | User | Command |
|-----------|-------|------|---------|
| Start SNMP (LAN) | Boot-up | root | `bash /usr/local/bin/start-snmpd.sh` |

---

## SNMP Metrics Reference (PromQL)

| Metric | Description |
|--------|-------------|
| `synology_system_status` | System health (1=Normal, 2=Failed) |
| `synology_temperature` | System temperature (Celsius) |
| `synology_power_status` | Power supply (1=Normal, 2=Failed) |
| `synology_system_fan_status` | Fan health (1=Normal, 2=Failed) |
| `synology_disk_status` | Per-disk status (1=Normal, 5=Crashed) |
| `synology_disk_temperature` | Per-disk temperature (Celsius) |
| `synology_disk_health_status` | Per-disk health (1=Normal, 4=Failing) |
| `synology_disk_bad_sector` | Per-disk bad sector count |
| `synology_raid_status` | Per-RAID status (1=Normal, 11=Degraded, 12=Crashed) |
| `synology_raid_free_size` | Per-RAID free bytes |
| `synology_raid_total_size` | Per-RAID total bytes |
| `synology_service_users` | Active users per service (CIFS, NFS, etc.) |
| `synology_storage_io_bytes_read` | Per-disk read bytes (counter) |
| `synology_storage_io_bytes_written` | Per-disk write bytes (counter) |
| `synology_space_io_bytes_read` | Per-volume read bytes (counter) |
| `synology_space_io_bytes_written` | Per-volume write bytes (counter) |

---

## Follow-ups

- [ ] Disable SSH password auth after confirming key auth is reliable
- [ ] Verify DSM boot task exists for SNMP (`bash /usr/local/bin/start-snmpd.sh`) and create it if missing
- [ ] Consider enabling DSM Auto Block (Control Panel → Security → Protection)
- [ ] Plan capacity expansion for volume2 (86%) and volume4 (88%)
- [ ] Monitor Disk 1 bad sectors (93 currently) — replace if count increases
- [ ] Set up Grafana alert rules for storage capacity and RAID degradation

---

## Server Inventory (Updated)

| Host | IP | Role | Metrics | Logs |
|------|-----|------|---------|------|
| sentinella.alpina | 172.16.19.94 | Observability | ✅ Self | ✅ Self |
| ntp.alpina | 172.16.16.108 | NTP Server | ✅ node_exporter | ✅ rsyslog |
| gotra | — | App Server | ✅ node_exporter | ✅ rsyslog |
| aria.alpina | 172.16.18.230 | Proxmox Host | ✅ node_exporter + PVE | ✅ rsyslog |
| pihole | 172.16.66.66 | DNS | ✅ node_exporter | ✅ rsyslog |
| komga.alpina | 172.16.16.202 | Media Server | ✅ node_exporter | — |
| home.alpina | 172.16.17.109 | Command Center | ✅ node_exporter | — |
| **portocali.alpina** | **172.16.21.21** | **NAS (77TB)** | **✅ SNMP exporter** | **✅ syslog-ng** |
