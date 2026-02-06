# IPv6 Preparation Plan — Alpina Homelab

**Status:** In Progress — Partially Working
**Date:** 2026-02-05 (updated from live audit)
**Goal:** Enable full dual-stack IPv6 across the entire Alpina homelab infrastructure

---

## Current State (Live Audit — 2026-02-05)

### Key Discovery: IPv6 Is Already Partially Working

A live audit of all 8 hosts revealed that **IPv6 is significantly further along than previously documented**. OPNsense has DHCPv6-PD active, Router Advertisements are being sent on LAN, and most hosts already have global IPv6 addresses via SLAAC.

### ISP Details

- **ISP:** Charter/Spectrum
- **Protocol:** DHCPv6 Prefix Delegation via `dhcp6c` on igb3 (WAN)
- **Delegated prefix:** `2603:8001:7400:fa9a::/64`
- **Previous (incorrect) documentation:** `2603:8001:7402:cf1c::/64` — this was stale/wrong
- **Current PD config:** Requesting a single `/64` (sla-id 0, sla-len 0)
- **Recommendation:** Request `/56` to enable subnetting (see Phase 1)
- Charter/Spectrum typically delegates `/56` prefixes, giving 256 `/64` subnets
- Prefixes are semi-stable (tied to DHCP lease) but not static — plan for prefix changes

### OPNsense WAN Config (`/var/etc/dhcp6c.conf`)

```
interface igb3 {
  send ia-pd 0;
  request domain-name-servers;
  request domain-name;
  script "/var/etc/dhcp6c_wan_script.sh";
};
id-assoc pd 0 {
  prefix-interface igb0 {
    sla-id 0;
    sla-len 0;
  };
};
```

### Per-Host IPv6 Status

| Host | IPv4 | IPv6 Address | Type | Ping6 | Issues |
|------|------|-------------|------|-------|--------|
| gateway.alpina (LAN) | 172.16.16.16 | `2603:8001:7400:fa9a::` range | Track IF | N/A | WAN has link-local only |
| pihole | 172.16.66.66 | `2603:8001:7400:fa9a:4392:b645:21ad:5510/64` | SLAAC+Privacy | Works | Privacy extensions = unstable addr |
| ntp.alpina | 172.16.16.108 | `2603:8001:7400:fa9a:be24:11ff:fe60:2dfe/64` | EUI-64 | Works | Stable (derived from MAC) |
| home.alpina | 172.16.17.109 | `2603:8001:7400:fa9a:be24:11ff:fec9:2694/64` | EUI-64 | **Timeout** | Firewalld blocks ICMPv6 outbound |
| komga.alpina | 172.16.16.202 | `2603:8001:7400:fa9a:be24:11ff:fe09:c0b9/64` | EUI-64 | Works | Docker bridge needs IPv6 config |
| sentinella.alpina | 172.16.19.94 | `2603:8001:7400:fa9a:be24:11ff:fe95:2956/64` | EUI-64 | **Timeout** | Firewalld blocks ICMPv6 outbound |
| aria.alpina (Proxmox) | 172.16.18.230 | **Link-local only** | None | **Unreachable** | No global IPv6; no /etc/network/interfaces config |
| homeassistant.local | 172.16.77.77 | **None visible** | N/A | **Unreachable** | SSH lands in container; host may have IPv6 |

### What's Working
- OPNsense DHCPv6-PD is active and receiving a prefix from Charter/Spectrum
- OPNsense has IPv6 forwarding enabled (`net.inet6.ip6.forwarding = 1`)
- OPNsense has 60 inet6 pf rules configured
- Router Advertisements are being sent on LAN (hosts are getting SLAAC addresses)
- 5 out of 8 hosts have global IPv6 addresses
- Pi-hole can resolve AAAA records (tested: `dig AAAA google.com @127.0.0.1` works)
- External IPv6 connectivity works from pihole, ntp, and komga
- NDP table shows many clients with `2603:8001:7400:fa9a::` addresses
- Multiple hosts see OPNsense as IPv6 default gateway

### What Needs Fixing
- [ ] **Prefix delegation size:** Currently `/64` — should request `/56` for future subnetting
- [ ] **Proxmox (aria):** No global IPv6 — needs `/etc/network/interfaces` updated
- [ ] **HAOS:** No visible IPv6 from SSH container — may need HA integration config
- [ ] **Firewall on home.alpina:** `firewalld` blocks outbound ICMPv6 — needs rule
- [ ] **Firewall on sentinella.alpina:** Same firewalld issue
- [ ] **Pi-hole privacy extensions:** Address changes on reboot — needs stable address for DNS
- [ ] **Pi-hole IPv6 upstreams:** Only using IPv4 upstream resolvers currently
- [ ] **DNS AAAA records:** No local AAAA records in Pi-hole `custom.list`
- [ ] **Monitoring:** No IPv6 panels in Grafana dashboard
- [ ] **Dual default routes:** Some hosts show two default routes (OPNsense + Pi-hole) — investigate

---

## Phase 1: OPNsense — Optimize Prefix Delegation

### 1.1 Request /56 Prefix (Currently Getting /64)

The current `sla-len 0` means OPNsense takes the full delegated prefix as a single /64. To enable subnetting for future VLANs:

**Navigate to:** Interfaces > WAN (igb3)

| Setting | Current | Recommended |
|---------|---------|-------------|
| IPv6 Configuration Type | DHCPv6 | DHCPv6 (no change) |
| Prefix delegation size | /64 (implicit) | **/56** |
| Send IPv6 prefix hint | Unknown | **Checked** |
| Request only an IPv6 prefix | Unknown | **Unchecked** (want both WAN addr + PD) |

If Charter/Spectrum rejects `/56`, fall back to `/60` (16 subnets). Never request just `/64` for PD — a `/64` cannot be subdivided.

**Verify after change:**
```bash
ssh root@172.16.16.16
cat /tmp/dhcp6c_info_igb3       # Shows delegated prefix
ifconfig igb0 | grep inet6      # LAN should get <prefix>::1
ping6 -c4 google.com            # External connectivity
```

### 1.2 LAN Track Interface (Already Working)

LAN igb0 is already configured as Track Interface tracking WAN. After upgrading to /56:

| Setting | Value |
|---------|-------|
| IPv6 Configuration Type | Track Interface |
| Track IPv6 Interface | WAN (igb3) |
| IPv6 Prefix ID | `0` (first /64 from the /56) |
| IPv6 Interface ID | `::1` (or `::16` to echo IPv4) |

If ISP delegates `2603:8001:7400:fa00::/56`, LAN gets `2603:8001:7400:fa00::/64`, gateway is `2603:8001:7400:fa00::1`.

### 1.3 Router Advertisements — Upgrade to Assisted Mode

**Current state:** RA is active (hosts get SLAAC addresses). Recommend upgrading to Assisted mode to push DNS.

**Navigate to:** Services > Router Advertisements > LAN

| Setting | Recommended |
|---------|-------------|
| RA Mode | **Assisted** (SLAAC + stateless DHCPv6) |
| Router Priority | High |
| Advertise Default Gateway | Checked |
| DNS servers | Pi-hole's stable IPv6 address (see Phase 2) |
| Domain search list | `alpina` |

**Why Assisted:** Addresses still assigned via SLAAC (simple, works everywhere), but the O-flag tells clients to also query DHCPv6 for DNS. This pushes Pi-hole's IPv6 address to all clients. RDNSS in RA alone is unreliable on some devices.

### 1.4 DHCPv6 Server (Stateless, for DNS Only)

**Navigate to:** Services > DHCPv6 > LAN

| Setting | Value |
|---------|-------|
| Enable | Checked |
| Mode | Stateless (no address assignment — SLAAC handles that) |
| DNS servers | Pi-hole's stable IPv6 address |

### 1.5 IPv6 Firewall Rules (Partially Done)

OPNsense already has 60 inet6 pf rules. Verify these critical rules exist:

**WAN inbound — Required ICMPv6:**

| ICMPv6 Type | Name | Why Required |
|-------------|------|-------------|
| 1 | Destination Unreachable | Error signaling |
| 2 | Packet Too Big | **Critical** — PMTUD. Blocking this breaks IPv6. |
| 3 | Time Exceeded | Traceroute, loop detection |
| 4 | Parameter Problem | Malformed packet notification |
| 128 | Echo Request | Ping (optional but recommended) |
| 129 | Echo Reply | Ping response |

**LAN — Required NDP ICMPv6 (MUST allow on LAN):**

| ICMPv6 Type | Name | Function |
|-------------|------|----------|
| 133 | Router Solicitation | Host asks for routers |
| 134 | Router Advertisement | Router sends prefix info |
| 135 | Neighbor Solicitation | IPv6 ARP request |
| 136 | Neighbor Advertisement | IPv6 ARP reply |
| 137 | Redirect | Better next-hop notification |

**Blocking types 133-137 completely breaks IPv6 addressing.** This is the most common IPv6 firewall mistake.

**LAN outbound:** Allow all IPv6 from LAN net (match IPv4 policy). Note: pf treats IPv4 and IPv6 as separate rulesets — IPv4 "allow all" does NOT cover IPv6.

**No NAT:** IPv6 has no NAT. Every host gets a globally routable address. WAN firewall is the only protection.

---

## Phase 2: Pi-hole DNS — IPv6 Configuration

### 2.1 Assign Stable IPv6 Address

Pi-hole currently uses SLAAC with **privacy extensions** (`4392:b645:21ad:5510` — random, changes on reboot). For DNS, it needs a stable address.

**Option A (Recommended): Disable privacy extensions on Pi-hole**
```bash
ssh -i ~/.ssh/id_ed25519 pi@pihole

# Disable temporary addresses for the primary interface
echo "net.ipv6.conf.eth0.use_tempaddr=0" | sudo tee /etc/sysctl.d/50-ipv6-stable.conf
sudo sysctl -p /etc/sysctl.d/50-ipv6-stable.conf

# Pi-hole will use the EUI-64 address derived from its MAC
# This is stable across reboots
ip -6 addr show eth0 | grep "scope global"
```

**Option B: Static IPv6 address**
```bash
# Add to /etc/network/interfaces or /etc/dhcpcd.conf:
# Static: 2603:8001:7400:fa9a::66/64  (matching IPv4 last octet)
```

### 2.2 Pi-hole v6 TOML Configuration

Pi-hole v6 uses `/etc/pihole/pihole.toml` (not the legacy `setupVars.conf`).

**IPv6 listening:** Pi-hole v6 automatically binds to both IPv4 and IPv6 when the host has an IPv6 address. No separate toggle needed.

```toml
# /etc/pihole/pihole.toml — relevant settings

[dns]
  # Listen on all local interfaces (IPv4 + IPv6 automatically)
  listeningMode = "local"

  # Add IPv6 upstream resolvers alongside existing IPv4 ones
  upstreams = [
    "9.9.9.11",          # Quad9 IPv4
    "1.1.1.1",           # Cloudflare IPv4
    "8.8.8.8",           # Google IPv4
    "2620:fe::11",       # Quad9 IPv6
    "2606:4700:4700::1111",  # Cloudflare IPv6
    "2001:4860:4860::8888"   # Google IPv6
  ]

  # "null" returns 0.0.0.0 for A and :: for AAAA on blocked domains
  blockingmode = "null"
```

**Verify after config change:**
```bash
sudo pihole restartdns
ss -tlnp | grep ':53'     # Should show 0.0.0.0:53 AND [::]:53
dig AAAA google.com @::1  # Test IPv6 DNS resolution
```

### 2.3 AAAA Records in Pi-hole

Add to `/etc/pihole/custom.list` (hosts-file format — Pi-hole auto-detects A vs AAAA):

```
# IPv6 AAAA records (add alongside existing IPv4 A records)
# Use actual addresses from live audit; update if prefix changes
2603:8001:7400:fa9a::1           gateway.alpina
2603:8001:7400:fa9a::66          pihole.alpina
2603:8001:7400:fa9a:be24:11ff:fe60:2dfe  ntp.alpina
2603:8001:7400:fa9a:be24:11ff:fe09:c0b9  komga.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956  sentinella.alpina
```

After editing: `sudo pihole restartdns`

**Note:** If the prefix changes (ISP lease renewal), all AAAA records must be updated. Consider using EUI-64 addresses (stable per-host) and only updating the prefix portion.

---

## Phase 3: Fix Hosts With Issues

### 3.1 Proxmox (aria.alpina) — No Global IPv6

Proxmox has only link-local on vmbr0. The `/etc/network/interfaces` has no IPv6 config, and IPv6 forwarding is commented out in `sysctl.conf`.

**Fix `/etc/network/interfaces`:**
```bash
ssh root@aria.alpina

# Add to the vmbr0 stanza in /etc/network/interfaces:
# (Do NOT set a static address — use SLAAC)
# Just ensure inet6 is not disabled

# Check current sysctl
sysctl net.ipv6.conf.vmbr0.accept_ra
# If forwarding=1 (Proxmox routes between VMs), must set accept_ra=2:
echo "net.ipv6.conf.vmbr0.accept_ra=2" >> /etc/sysctl.d/50-ipv6.conf
sysctl -p /etc/sysctl.d/50-ipv6.conf

# Verify
ip -6 addr show vmbr0   # Should now get global address
ping6 -c4 google.com
```

**Important:** On Linux, when `forwarding=1`, the kernel ignores Router Advertisements unless `accept_ra=2`. This is the most likely reason Proxmox has no global IPv6.

### 3.2 home.alpina — Firewalld Blocks ICMPv6

Has global IPv6 (`2603:8001:7400:fa9a:be24:11ff:fec9:2694/64`) but ping6 times out due to `firewalld` only allowing ssh+http.

```bash
ssh -i ~/.ssh/id_ed25519 alfa@home.alpina

# Allow ICMPv6 (required for IPv6 to work properly)
sudo firewall-cmd --permanent --add-protocol=ipv6-icmp
sudo firewall-cmd --reload

# Verify
ping6 -c4 google.com
```

### 3.3 sentinella.alpina — Firewalld Blocks ICMPv6

Same issue as home.alpina.

```bash
ssh alfa@sentinella.alpina

sudo firewall-cmd --permanent --add-protocol=ipv6-icmp
sudo firewall-cmd --reload

ping6 -c4 google.com
```

### 3.4 Home Assistant (HAOS) — Limited Access

SSH lands inside a containerized addon, not the host OS. The host may have IPv6 via SLAAC, but we can't verify from the SSH container.

**Options:**
- Check HA web UI for network settings (Settings > System > Network)
- HAOS should receive SLAAC address automatically if the host OS supports it
- No action needed unless IPv6 access to HA is specifically required

### 3.5 Dual Default Routes

Some hosts show two IPv6 default routes (OPNsense and Pi-hole as next-hops). This happens when Pi-hole's radvd or router advertisement daemon is inadvertently advertising itself as a router.

**Investigation:**
```bash
# On any affected host:
ip -6 route show default
# Look for multiple "via fe80::..." entries

# On Pi-hole — check if radvd is running:
ssh -i ~/.ssh/id_ed25519 pi@pihole 'ps aux | grep -i radvd'
```

If Pi-hole is sending RAs, disable it — only OPNsense should be the IPv6 gateway.

---

## Phase 4: Server Service Binding

### 4.1 node_exporter (All Hosts)

node_exporter typically binds to `:::9100` (dual-stack) by default. Verify on each host:

```bash
ss -tlnp | grep 9100
# Want: [::]:9100  (dual-stack)
# Bad:  0.0.0.0:9100  (IPv4 only)
```

If IPv4-only, restart with `--web.listen-address=:9100` (colon prefix = dual-stack).

### 4.2 Chrony (ntp.alpina)

Check if chronyd binds to IPv6:
```bash
ssh -i ~/.ssh/id_ed25519 alfa@ntp.alpina
grep "bindaddress\|allow" /etc/chrony.conf
ss -ulnp | grep chronyd
```

If not binding IPv6, add to `/etc/chrony.conf`:
```
bindaddress ::
allow 2603:8001:7400:fa9a::/64
```

### 4.3 Komga Docker (komga.alpina)

Docker's default bridge network does NOT support IPv6. To enable:
```bash
ssh -i ~/.ssh/id_ed25519_komga_alpina alfa@komga.alpina

# Check current Docker IPv6 support
docker network inspect bridge | grep EnableIPv6

# If false, add to /etc/docker/daemon.json:
# {
#   "ipv6": true,
#   "fixed-cidr-v6": "fd00:dead:beef::/48"
# }
# Then: sudo systemctl restart docker
```

### 4.4 Sentinella Podman Stack

Podman containers need port bindings that include IPv6. Check Caddy (reverse proxy) config:
```bash
ssh alfa@sentinella.alpina
sudo podman inspect <caddy-container> | grep -A5 HostPort
```

---

## Phase 5: Monitoring & Observability

### 5.1 Prometheus IPv6 Targets

Once hosts have stable IPv6, optionally add IPv6 targets. The simplest approach is using hostnames that resolve to both A and AAAA records — Prometheus will use the hostname.

Current targets in `/opt/observability/prometheus/prometheus.yml` use hostnames, which will automatically resolve to IPv6 once AAAA records are added to Pi-hole.

### 5.2 Grafana Dashboard Panels

Add to the Alpina Homelab Command Center:
- [ ] IPv6 connectivity indicator per host
- [ ] IPv6 vs IPv4 traffic ratio (if OPNsense exports per-protocol stats)
- [ ] ICMPv6 RA status panel

### 5.3 Syslog over IPv6

Alloy syslog receiver already listens on `:::1514`. Once hosts have IPv6, syslog forwarding will work over IPv6 automatically if rsyslog configs use hostnames.

---

## Phase 6: Testing & Validation

### 6.1 Per-Host Connectivity

```bash
# From each host:
ping6 -c4 google.com                         # External IPv6
ping6 -c4 2603:8001:7400:fa9a::1             # Gateway
ping6 -c4 <pihole-ipv6>                      # DNS server
curl -6 https://ipv6.google.com              # HTTPS over IPv6
```

### 6.2 DNS Resolution

```bash
dig AAAA google.com @<pihole-ipv6>           # External AAAA
dig AAAA ntp.alpina @<pihole-ipv6>           # Internal AAAA
dig AAAA komga.alpina @<pihole-ipv6>         # Internal AAAA
```

### 6.3 Service Access Over IPv6

```bash
curl -6 https://grafana.sentinella.alpina    # Grafana
curl -6 http://ntp.alpina                    # NTP landing page
curl -6 http://komga.alpina                  # Komga landing page
curl -6 http://komga.alpina:25600            # Komga UI
```

---

## Suggested IPv6 Address Plan

Using stable addresses for servers (EUI-64 from MAC or static assignment). Last octets mirror IPv4 where practical.

| Host | IPv4 | Suggested Stable IPv6 | Current Live IPv6 |
|------|------|-----------------------|-------------------|
| gateway.alpina | 172.16.16.16 | `<prefix>::1` | (LAN gateway) |
| pihole | 172.16.66.66 | `<prefix>::66` (static) | SLAAC+privacy (unstable) |
| ntp.alpina | 172.16.16.108 | `<prefix>::108` or keep EUI-64 | `...fa9a:be24:11ff:fe60:2dfe` |
| komga.alpina | 172.16.16.202 | `<prefix>::202` or keep EUI-64 | `...fa9a:be24:11ff:fe09:c0b9` |
| sentinella.alpina | 172.16.19.94 | `<prefix>::94` or keep EUI-64 | `...fa9a:be24:11ff:fe95:2956` |
| home.alpina | 172.16.17.109 | `<prefix>::109` or keep EUI-64 | `...fa9a:be24:11ff:fec9:2694` |
| aria.alpina | 172.16.18.230 | `<prefix>::230` or keep EUI-64 | **No global IPv6 yet** |
| homeassistant.local | 172.16.77.77 | `<prefix>::77` | **Unknown** |

**Note:** EUI-64 addresses are stable (derived from MAC) but long. Static addresses (e.g., `::66`) are shorter and easier to remember. Either approach works — EUI-64 requires no additional configuration, static requires manual setup on each host.

**Prefix change risk:** If `<prefix>` changes (ISP lease renewal), all static addresses change too. EUI-64 host portions stay the same but the prefix still changes. Use hostnames in configs, not hardcoded IPv6 addresses.

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| ISP changes delegated prefix | All IPv6 addresses change | Use hostnames everywhere; script AAAA record updates |
| Broken ICMPv6 filtering | IPv6 stops working entirely | Always allow ICMPv6 types 1-4, 128-137 |
| Privacy extensions change addresses | DNS records stale | Use static IPs or disable tempaddr on servers |
| Application only binds to IPv4 | Service unreachable on IPv6 | Verify `:::port` binding on each service |
| IPv6 bypasses Pi-hole ad blocking | Ads return | Ensure Pi-hole is the only IPv6 DNS server via RA/DHCPv6 |
| Proxmox forwarding + accept_ra | No SLAAC addresses | Set `accept_ra=2` when `forwarding=1` |
| Docker/Podman no IPv6 | Containerized services unreachable | Enable IPv6 in container runtime config |
| Android devices ignore DHCPv6 | No DNS via DHCPv6 | Use RDNSS in RA as backup (both methods simultaneously) |

---

## Rollback Plan

If IPv6 causes issues:
1. Disable RA on OPNsense LAN interface (stops IPv6 address assignment)
2. Remove IPv6 firewall rules
3. Remove AAAA records from Pi-hole
4. Hosts fall back to IPv4 automatically (dual-stack)

**IPv6 on the WAN side can remain enabled** even if LAN IPv6 is disabled.

---

## Order of Operations

| Step | Phase | Task | Status |
|------|-------|------|--------|
| 1 | 1.1 | Upgrade PD request from /64 to /56 | Todo |
| 2 | 1.3 | Upgrade RA mode to Assisted (push DNS) | Todo |
| 3 | 1.5 | Verify IPv6 firewall rules on OPNsense | Partially done (60 rules exist) |
| 4 | 2.1 | Assign Pi-hole a stable IPv6 address | Todo |
| 5 | 2.2 | Add IPv6 upstream DNS resolvers to pihole.toml | Todo |
| 6 | 2.3 | Add AAAA records to Pi-hole custom.list | Todo |
| 7 | 3.1 | Fix Proxmox — enable accept_ra=2 on vmbr0 | Todo |
| 8 | 3.2 | Fix home.alpina — allow ICMPv6 in firewalld | Todo |
| 9 | 3.3 | Fix sentinella.alpina — allow ICMPv6 in firewalld | Todo |
| 10 | 3.5 | Investigate dual default routes | Todo |
| 11 | 4 | Verify service binding on all hosts | Todo |
| 12 | 5 | Add IPv6 monitoring panels to Grafana | Todo |
| 13 | 6 | End-to-end testing | Todo |

---

## References

- [OPNsense IPv6 Configuration](https://docs.opnsense.org/manual/ipv6.html)
- [OPNsense Router Advertisements](https://docs.opnsense.org/manual/radvd.html)
- [Pi-hole v6 pihole.toml Reference](https://docs.pi-hole.net/core/pihole-toml/)
- [Charter/Spectrum IPv6 Info](https://www.spectrum.net/support/internet/ipv6)
- [RFC 8415 — DHCPv6 (includes Prefix Delegation)](https://www.rfc-editor.org/rfc/rfc8415)
- [RFC 4861 — Neighbor Discovery for IPv6](https://www.rfc-editor.org/rfc/rfc4861)
- [RFC 4862 — IPv6 SLAAC](https://www.rfc-editor.org/rfc/rfc4862)
- [RFC 8106 — RDNSS/DNSSL in RA](https://www.rfc-editor.org/rfc/rfc8106)
- [RFC 4443 — ICMPv6 (Packet Too Big must never be blocked)](https://www.rfc-editor.org/rfc/rfc4443)
