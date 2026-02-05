# NTP Server Security Audit & Improvement Plan

**Server:** ntp.alpina (172.16.16.108)
**OS:** AlmaLinux 10.1 (Heliotrope Lion)
**Kernel:** 6.12.0-124.21.1.el10_1.x86_64
**NTP Service:** chrony 4.6.1
**Assessment Date:** 2026-01-27
**Auditor:** Automated Assessment

---

## Executive Summary

The NTP server is **operational and functional** but requires hardening and operational improvements to meet professional standards. The server correctly synchronizes time with upstream sources and serves clients on the internal network, but lacks monitoring, comprehensive logging, diverse time sources, and documentation.

| Category | Status | Grade |
|----------|--------|-------|
| Time Synchronization | ✅ Working | A |
| Security Hardening | ⚠️ Partial | C |
| Monitoring & Logging | ❌ Missing | F |
| Source Diversity | ⚠️ Partial | C |
| Documentation | ❌ Missing | F |
| **Overall** | **Functional but needs work** | **C+** |

---

## Server Inventory

### Hardware Profile

| Component | Specification |
|-----------|--------------|
| CPU | Intel N100 (1 core) |
| Memory | 1.7 GB total, ~1.3 GB available |
| Storage | 444 GB disk (LVM: 70G root, 2G swap, 371G home) |
| Virtualization | KVM guest (QEMU) |
| Network | Single NIC (ens18), 1500 MTU |

### Network Configuration

| Interface | IPv4 | IPv6 | Gateway |
|-----------|------|------|---------|
| ens18 | 172.16.16.108/16 (DHCP) | 2603:8001:7402:cf1c:be24:11ff:fe60:2dfe/64 | 172.16.16.16 |

**DNS:** 172.16.66.66 (alpina domain)

### Running Services

| Service | State | Notes |
|---------|-------|-------|
| chronyd | Running (20d uptime) | NTP server |
| firewalld | Running | Active, config unknown |
| sshd | Running | Port 22 |
| cockpit | Enabled | Port 9090 (web console) |
| rsyslog | Running | System logging |
| tuned | Running | Profile: virtual-guest |
| auditd | Running | Security auditing |

### Listening Ports

| Port | Protocol | Service | Binding |
|------|----------|---------|---------|
| 22 | TCP | SSH | 0.0.0.0 / :: |
| 123 | UDP | NTP | 0.0.0.0 (IPv4 only) |
| 323 | UDP | Chrony cmdmon | 127.0.0.1 / ::1 |
| 9090 | TCP | Cockpit | * |

### User Accounts

| User | UID | Groups | Shell | Notes |
|------|-----|--------|-------|-------|
| root | 0 | root | /bin/bash | System |
| alfa | 1000 | alfa, wheel | /bin/bash | Admin user, sudo via wheel |

---

## Current NTP Configuration

### /etc/chrony.conf

```conf
# Upstream NTP servers (US pool)
server 0.us.pool.ntp.org iburst
server 1.us.pool.ntp.org iburst
server 2.us.pool.ntp.org iburst
server 3.us.pool.ntp.org iburst

# Common baseline settings
driftfile /var/lib/chrony/drift
makestep 1.0 3
rtcsync
keyfile /etc/chrony.keys
leapsectz right/UTC
logdir /var/log/chrony
cmdallow 192.168.1.0/24

allow 172.16.16.0/24
allow 192.168.1.0/24
```

### Chrony Service Options

```
OPTIONS="-F 2"  # seccomp filter level 2 (hardened)
```

### Current Time Sources

| Source | Stratum | Status | Offset | Freq Error |
|--------|---------|--------|--------|------------|
| ntp.shastacoe.net | 2 | Combined (+) | +2.7ms | +0.148 ppm |
| rls-clock01.rlasd.net | 1 | Combined (+) | -9.2ms | -0.031 ppm |
| ntp.maxhost.io | 2 | Combined (+) | +5.8ms | +0.434 ppm |
| 136-55-44-51.googlefiber.net | 2 | **Selected (*)** | +3.3ms | +0.250 ppm |

### Sync Status

| Metric | Value | Assessment |
|--------|-------|------------|
| Reference | 136-55-44-51.googlefiber.net | Good |
| Stratum | 3 | Acceptable |
| System Time Offset | +1.1ms | Excellent |
| Root Delay | 59.4ms | Normal |
| Frequency Error | 25.2 ppm slow | Normal |
| Leap Status | Normal | Good |

---

## Security Assessment

### ✅ Strengths

1. **SELinux Enforcing** - Mandatory access controls active
2. **Systemd Hardening** - Chrony runs with capability restrictions:
   - PrivateTmp, ProtectHome, ProtectSystem=strict
   - MemoryDenyWriteExecute, NoNewPrivileges
   - RestrictNamespaces, RestrictAddressFamilies
3. **Seccomp Filtering** - `-F 2` flag enables seccomp sandbox
4. **Command Socket Restricted** - Only localhost can run chronyc
5. **Wheel Group Sudo** - Proper privilege separation
6. **Audit Daemon Running** - Security logging available
7. **Firewall Active** - firewalld running (rules need verification)

### ❌ Vulnerabilities & Gaps

| Issue | Severity | Risk |
|-------|----------|------|
| No NTP authentication | Medium | Clients could sync to rogue server |
| Empty chrony.keys file | Medium | Authentication not configured |
| Logging inaccessible | High | Cannot audit or troubleshoot |
| No monitoring/alerting | High | Failures go undetected |
| Cockpit exposed (9090) | Low | Web management interface accessible |
| No automatic updates | Medium | Security patches not applied |
| IPv6 NTP not enabled | Low | IPv6 clients cannot sync |
| Pool-only sources | Medium | No stratum 1 or vendor diversity |
| cmdallow too broad | Low | 192.168.1.0/24 can run chronyc |
| DHCP-assigned IP | Low | NTP server IP could change |

### Unknown (Requires Sudo)

- Firewall rules and zones
- SSH configuration details
- Full audit logs
- Chrony drift file contents

---

## Findings Summary

### Critical (Fix Immediately)

1. **Enable Logging**
   - Log directory exists but is inaccessible
   - No visibility into client activity or errors
   - Add: `log measurements statistics tracking`

2. **Implement Monitoring**
   - No health checks for time drift
   - No alerting for service failures
   - Create systemd timer for health checks

### High Priority

3. **Diversify Time Sources**
   - Currently using only pool.ntp.org
   - Add stratum 1 servers (NIST, Google, Cloudflare)
   - Ensure geographic and organizational diversity

4. **Verify Firewall Rules**
   - Confirm NTP service is properly allowed
   - Restrict to internal networks only
   - Block external access if not needed

5. **Configure Automatic Updates**
   - dnf-automatic not enabled
   - Security patches not being applied
   - Enable at minimum security updates

### Medium Priority

6. **Enable NTP Authentication**
   - Generate keys for chrony.keys
   - Require authentication for sensitive clients
   - Document key distribution process

7. **Add Rate Limiting**
   - No DDoS protection configured
   - Add: `ratelimit interval 1 burst 16 leak 2`

8. **Enable IPv6 NTP Service**
   - Server has IPv6 address
   - NTP only listening on IPv4
   - Add: `bindaddress ::`

9. **Static IP Configuration**
   - DHCP lease could cause IP change
   - NTP servers should have static IPs
   - Configure static IP or DHCP reservation

10. **Review Cockpit Access**
    - Web console on port 9090
    - Consider restricting to localhost or VPN

### Low Priority

11. **Create Documentation**
    - No runbook exists
    - No recovery procedures
    - No network diagram

12. **Configuration Backup**
    - No automated backups
    - Create daily backup script

13. **Redundant NTP Server**
    - Single point of failure
    - Deploy secondary NTP server

---

## Remediation Plan

### Phase 1: Critical Fixes (Day 1)

```bash
# 1. Fix logging (requires sudo)
sudo mkdir -p /var/log/chrony
sudo chown chrony:chrony /var/log/chrony
sudo chmod 750 /var/log/chrony

# Add to /etc/chrony.conf:
# log measurements statistics tracking rtc refclocks
# logchange 0.1
# clientlog

sudo systemctl restart chronyd
```

### Phase 2: Security Hardening (Day 1-2)

```bash
# 1. Verify and configure firewall
sudo firewall-cmd --list-all
sudo firewall-cmd --permanent --add-service=ntp
sudo firewall-cmd --reload

# 2. Add rate limiting to /etc/chrony.conf:
# ratelimit interval 1 burst 16 leak 2
# clientloglimit 10000

# 3. Enable automatic security updates
sudo dnf install dnf-automatic
sudo systemctl enable --now dnf-automatic-install.timer
```

### Phase 3: Source Diversity (Day 2)

Replace pool servers in /etc/chrony.conf:

```conf
# Primary - Stratum 1 and major providers
server time.nist.gov iburst minpoll 4 maxpoll 10
server time.google.com iburst prefer minpoll 4 maxpoll 10
server time.cloudflare.com iburst minpoll 4 maxpoll 10

# Secondary - Additional diversity
server time1.google.com iburst minpoll 4 maxpoll 10
server time2.google.com iburst minpoll 4 maxpoll 10

# Fallback - Pool with limits
pool us.pool.ntp.org iburst maxsources 2 minpoll 4 maxpoll 10

# Require minimum sources
minsources 3
```

### Phase 4: Monitoring (Day 2-3)

Create `/usr/local/bin/check-ntp-health.sh`:

```bash
#!/bin/bash
# NTP Health Check

OFFSET_THRESHOLD_MS=10
MIN_SOURCES=3

# Check service
if ! systemctl is-active --quiet chronyd; then
    echo "CRITICAL: chronyd not running"
    exit 2
fi

# Check offset
offset=$(chronyc tracking | awk '/System time/{print $4}')
offset_ms=$(echo "$offset * 1000" | bc | cut -d. -f1)
offset_ms=${offset_ms#-}

if [ "$offset_ms" -gt "$OFFSET_THRESHOLD_MS" ]; then
    echo "WARNING: Offset ${offset}s exceeds ${OFFSET_THRESHOLD_MS}ms"
    exit 1
fi

# Check sources
sources=$(chronyc sources | grep -c "^\^[*+]")
if [ "$sources" -lt "$MIN_SOURCES" ]; then
    echo "WARNING: Only $sources good sources (need $MIN_SOURCES)"
    exit 1
fi

echo "OK: offset=${offset}s, sources=$sources"
exit 0
```

Enable with systemd timer running every 5 minutes.

### Phase 5: IPv6 & Cleanup (Day 3)

```bash
# Add to /etc/chrony.conf:
# bindaddress 0.0.0.0
# bindaddress ::
# allow 2603:8001:7402:cf1c::/64

# Configure static IP (via nmcli or Cockpit)
# Update firewall for IPv6 NTP
```

### Phase 6: Documentation (Week 1)

Create:
- Server runbook with procedures
- Network diagram showing NTP clients
- Recovery procedures
- Change management process

---

## Recommended Configuration

### /etc/chrony.conf (Production)

```conf
#===============================================
# NTP Server Configuration - ntp.alpina
# Last Modified: YYYY-MM-DD
#===============================================

#-----------------------------------------------
# Upstream Time Sources
#-----------------------------------------------
# Stratum 1 - NIST
server time.nist.gov iburst minpoll 4 maxpoll 10

# Anycast - Google (preferred)
server time.google.com iburst prefer minpoll 4 maxpoll 10
server time1.google.com iburst minpoll 4 maxpoll 10
server time2.google.com iburst minpoll 4 maxpoll 10

# Anycast - Cloudflare
server time.cloudflare.com iburst minpoll 4 maxpoll 10

# Pool fallback
pool us.pool.ntp.org iburst maxsources 2 minpoll 4 maxpoll 10

# Minimum sources before declaring sync
minsources 3

#-----------------------------------------------
# Clock Discipline
#-----------------------------------------------
driftfile /var/lib/chrony/drift
makestep 1.0 3
maxupdateskew 100.0
rtcsync

#-----------------------------------------------
# Leap Second Handling
#-----------------------------------------------
leapsectz right/UTC
leapsecmode slew
maxslewrate 1000.0

#-----------------------------------------------
# Local Fallback (if all sources lost)
#-----------------------------------------------
local stratum 10 orphan

#-----------------------------------------------
# Client Access Control
#-----------------------------------------------
# IPv4 networks
allow 172.16.0.0/16
allow 192.168.1.0/24

# IPv6 networks
allow 2603:8001:7402::/48
allow fe80::/10

# DDoS protection
ratelimit interval 1 burst 16 leak 2
clientloglimit 10000

#-----------------------------------------------
# Administrative Access
#-----------------------------------------------
bindcmdaddress 127.0.0.1
bindcmdaddress ::1
cmdallow 127.0.0.1
cmdratelimit interval 3 burst 1 leak 4

#-----------------------------------------------
# Authentication
#-----------------------------------------------
keyfile /etc/chrony.keys
# Uncomment after configuring keys:
# cmdkey 1

#-----------------------------------------------
# Logging
#-----------------------------------------------
logdir /var/log/chrony
log measurements statistics tracking rtc
logchange 0.1
logbanner 64
clientlog

#-----------------------------------------------
# Network Binding
#-----------------------------------------------
bindaddress 0.0.0.0
bindaddress ::
```

### /etc/logrotate.d/chrony

```
/var/log/chrony/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    sharedscripts
    postrotate
        /usr/bin/chronyc cyclelogs > /dev/null 2>&1 || true
    endscript
}
```

---

## Verification Commands

After implementing changes, verify with:

```bash
# Service status
systemctl status chronyd

# Source sync
chronyc sources -v
chronyc sourcestats

# Time tracking
chronyc tracking

# Client activity (after enabling clientlog)
chronyc clients

# Port binding
ss -ulnp | grep 123

# Firewall
sudo firewall-cmd --list-all

# Test from client
ntpdate -q 172.16.16.108
```

---

## Success Criteria

| Metric | Target | Verification |
|--------|--------|--------------|
| Service uptime | >99.9% | `systemctl status chronyd` |
| Time offset | <5ms | `chronyc tracking` |
| Sources available | ≥4 | `chronyc sources` |
| Good sources | ≥3 | `chronyc sources \| grep "^\^[*+]"` |
| Logging enabled | Active | `ls /var/log/chrony/` |
| Monitoring | Running | `systemctl status ntp-health-check.timer` |
| IPv6 service | Listening | `ss -ulnp \| grep ":::123"` |

---

## Appendix: Quick Reference

### Chronyc Commands

```bash
chronyc sources -v      # View time sources
chronyc sourcestats     # Source statistics
chronyc tracking        # Sync status
chronyc clients         # Client list (if clientlog enabled)
chronyc activity        # Online/offline sources
chronyc serverstats     # Server statistics (requires auth)
chronyc makestep        # Force time step (emergency)
```

### Troubleshooting

```bash
# Restart service
sudo systemctl restart chronyd

# Check logs
journalctl -u chronyd -f

# Force source poll
chronyc burst 1/1

# Check connectivity to source
ping time.google.com
traceroute time.nist.gov
```

### Emergency Procedures

```bash
# If time is way off (>1 second):
sudo chronyc makestep

# If no sources available:
# 1. Check network: ping time.google.com
# 2. Check firewall: sudo firewall-cmd --list-all
# 3. Check DNS: dig time.google.com
# 4. Manual sync: sudo chronyd -q 'server time.google.com iburst'

# If service won't start:
sudo chronyd -d  # Debug mode
journalctl -u chronyd -b
```

---

## Document History

| Date | Author | Changes |
|------|--------|---------|
| 2026-01-27 | Automated Audit | Initial assessment |

---

**END OF DOCUMENT**
