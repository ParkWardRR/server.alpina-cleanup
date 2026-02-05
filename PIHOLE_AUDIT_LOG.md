# Pi-hole Audit & Maintenance Log

## System Details
- **Host:** pihole (172.16.66.66)
- **OS:** Debian GNU/Linux 12 (bookworm)
- **Hardware:** Raspberry Pi (arm64)
- **Domain:** alpina

---

## 2026-01-27 - Initial Audit & System Update

### Actions Completed

#### 1. System Package Updates
- **Status:** COMPLETED
- **Packages Updated:** 195 packages upgraded, 16 newly installed
- **Notable Updates:**
  - `bash` 5.2.15-2+b7 → 5.2.15-2+b10
  - `curl` 7.88.1-10+deb12u8 → 7.88.1-10+deb12u14
  - `openssh-server` updated
  - `bind9-libs` updated (security)
  - `dnsmasq-base` 2.89-1 → 2.90-4~deb12u1
  - Linux kernel 6.6.51 → 6.12.62
- **Note:** Required manual fix for `initramfs-tools-core` config prompt (kept existing config)

#### 2. Pi-hole Update
- **Status:** COMPLETED (already at latest)
- **Versions:**
  - Core: v6.3 (Latest)
  - Web: v6.4 (Latest)
  - FTL: v6.4.1 (Latest)

#### 3. Gravity/Blocklist Check
- **Status:** WORKING
- **Active Adlists:**
  - `https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts` (71,934 domains)
  - `https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt` (34 domains)
- **Last Update:** 2026-01-25 04:49

---

### Issues Identified (Initial Scan)

| Issue | Severity | Status | Notes |
|-------|----------|--------|-------|
| 195 packages outdated | CRITICAL | FIXED | All packages updated |
| SSH password auth enabled | CRITICAL | SKIPPED | User deferred |
| DNS listening mode = ALL | HIGH | SKIPPED | User deferred |
| Too many upstream DNS servers (12) | MEDIUM | SKIPPED | User deferred |
| No fail2ban | MEDIUM | SKIPPED | User deferred |
| DNSSEC disabled | LOW | SKIPPED | User deferred |
| NetworkManager DNS override | MEDIUM | SKIPPED | User deferred |
| 2FA not enabled on web UI | LOW | SKIPPED | User deferred |
| Router DHCP configuration | MEDIUM | DONE | User confirmed |
| No reverse DNS configured | LOW | PENDING | - |
| IPv6 disabled | INFO | PLAN NEEDED | See below |

---

### Current Configuration Summary

#### DNS Settings
```toml
upstreams = [
  "9.9.9.11",           # Quad9 IPv4
  "149.112.112.11",     # Quad9 IPv4 alt
  "2620:fe::11",        # Quad9 IPv6
  "2620:fe::fe:11",     # Quad9 IPv6 alt
  "1.1.1.1",            # Cloudflare IPv4
  "1.0.0.1",            # Cloudflare IPv4 alt
  "2606:4700:4700::1111", # Cloudflare IPv6
  "2606:4700:4700::1001", # Cloudflare IPv6 alt
  "8.8.8.8",            # Google IPv4
  "8.8.4.4",            # Google IPv4 alt
  "2001:4860:4860::8888", # Google IPv6
  "2001:4860:4860::8844"  # Google IPv6 alt
]

listeningMode = "ALL"
dnssec = false
domain.name = "alpina"
blockESNI = true
mozillaCanary = true
iCloudPrivateRelay = true
```

#### Network
- **Interface:** eth0
- **IPv4:** 172.16.66.66/16
- **Gateway:** 172.16.16.16
- **IPv6:** DISABLED (only loopback ::1)

#### Services Running
- pihole-FTL (DNS + Web)
- sshd (port 22)
- NetworkManager
- avahi-daemon
- bluetooth

#### Custom DNS Records
```
172.16.66.66 pihole.alpina
172.16.20.30 zagato.alpina
172.16.21.21 portocali
172.16.21.69 munich
172.16.21.185 my.ampache.alpina
172.16.20.20 trofeo
172.16.20.20 engenius
172.16.19.77 meshcentral.alpina
172.16.18.230 aria
172.16.18.230 aria.alpina
172.16.20.57 ezmaster
172.16.16.108 ntp
```

---

## IPv6 Implementation Plan

### Current State
- IPv6 is **disabled** in NetworkManager (`method=disabled`)
- Pi-hole is listening on IPv6 for DNS (UDP/TCP on port 53)
- IPv6 upstream DNS servers are configured but unused (no local IPv6 address)
- Only loopback `::1` is present

### Prerequisites to Check (Manual)
1. **Does your ISP provide IPv6?**
   - Check router admin panel for IPv6 WAN status
   - Or run `curl -6 ifconfig.co` from a device with IPv6 enabled

2. **What type of IPv6 does your ISP provide?**
   - Native dual-stack (preferred)
   - 6rd tunnel
   - DS-Lite (may complicate things)

3. **Router IPv6 support?**
   - Check if router supports IPv6 DHCP (DHCPv6) or SLAAC
   - Router model: Unknown (at 172.16.16.16)

### Implementation Steps (Once Prerequisites Confirmed)

#### Step 1: Enable IPv6 on Router
- Enable IPv6 WAN connection (if ISP supports)
- Configure IPv6 LAN distribution (SLAAC or DHCPv6)
- Note the prefix delegated by ISP (e.g., `2001:db8:abcd::/48`)

#### Step 2: Configure Static IPv6 on Pi-hole
Option A - SLAAC with static suffix:
```bash
# Add to NetworkManager connection
sudo nmcli connection modify "Wired connection 1" ipv6.method "auto"
sudo nmcli connection modify "Wired connection 1" ipv6.addr-gen-mode "eui64"
```

Option B - Full static IPv6 (recommended for servers):
```bash
# Replace with your actual prefix
sudo nmcli connection modify "Wired connection 1" ipv6.method "manual"
sudo nmcli connection modify "Wired connection 1" ipv6.addresses "2001:db8:abcd::66/64"
sudo nmcli connection modify "Wired connection 1" ipv6.gateway "fe80::1"
sudo nmcli connection modify "Wired connection 1" ipv6.dns "::1"
```

#### Step 3: Update Pi-hole Configuration
```bash
# Verify Pi-hole listens on new IPv6 address
pihole status

# If needed, update pihole.toml to bind to specific interface
# dns.interface = "eth0"
```

#### Step 4: Update Router DHCP/SLAAC
- Configure router to advertise Pi-hole's IPv6 address as DNS server
- For DHCPv6: Add the IPv6 address to DNS server list
- For SLAAC: May need to use Router Advertisement options

#### Step 5: Test
```bash
# From Pi-hole
ping6 google.com
dig @::1 google.com AAAA

# From client device
dig @<pihole-ipv6> google.com AAAA
```

### Risks & Considerations
1. **Privacy:** IPv6 addresses can be more trackable (consider privacy extensions for clients)
2. **Complexity:** Dual-stack means maintaining two configurations
3. **Firewall:** Ensure router firewall covers IPv6 (no NAT protection like IPv4)
4. **Fallback:** If IPv6 breaks, clients should fallback to IPv4

### Recommendation
Start with **Option A (SLAAC)** to test IPv6 connectivity, then migrate to **Option B (static)** for production stability. This allows you to verify ISP IPv6 works before committing to a static configuration.

---

## What's Working Well
- Pi-hole v6 is current and blocking enabled
- 71,968 domains in blocklists
- TLS/HTTPS enabled on web interface
- Anti-bypass measures enabled (ESNI, Mozilla canary, iCloud Private Relay)
- Rate limiting configured (1000 queries/60s)
- Weekly gravity updates scheduled
- System uptime: 57 days (stable)
- Disk usage: 15% (healthy)
- Memory usage: 312MB of 907MB (healthy)

---

## Next Maintenance Actions
- [ ] Reboot to load new kernel (6.12.62) - **recommended after next maintenance window**
- [ ] Consider enabling DNSSEC for DNS validation
- [ ] Consider reducing upstream DNS servers to 2-4
- [ ] Consider hardening SSH (disable password auth)
- [ ] Consider installing fail2ban
- [ ] Implement IPv6 (see plan above)
- [ ] Configure reverse DNS/conditional forwarding for local hostname resolution

---

*Last updated: 2026-01-27 13:20 PST*
