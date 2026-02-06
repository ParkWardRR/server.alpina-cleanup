# IPv6 Preparation Plan — Alpina Homelab

**Status:** Planning
**Date:** 2026-02-05
**Goal:** Enable dual-stack IPv6 across the entire Alpina homelab infrastructure

---

## Current State

### What We Have
- **ISP:** Charter/Spectrum — provides IPv6 prefix via DHCPv6-PD
- **Known prefix:** `2603:8001:7402:cf1c::/64` (observed on OPNsense WAN)
- **OPNsense:** v23.7.12 on FreeBSD 13.2 — has full IPv6 support
- **All VMs:** Linux-based (AlmaLinux, Debian) — IPv6 kernel support built-in
- **Pi-hole:** Already handles DNS; can serve AAAA records

### What's Missing
- IPv6 is not actively configured on LAN interfaces
- No internal IPv6 addressing scheme
- No IPv6 firewall rules
- DNS (Pi-hole) not configured for IPv6 resolution
- No IPv6 monitoring in Grafana

---

## Phase 1: OPNsense Gateway Configuration

### 1.1 WAN IPv6 (DHCPv6-PD)
- [ ] Verify ISP provides DHCPv6 prefix delegation
- [ ] Configure WAN interface for DHCPv6
  - Interface: `igb3` (WAN)
  - Request prefix size: `/56` or `/60` (depends on ISP)
  - Enable "Send IPv6 prefix hint"
- [ ] Verify WAN gets a global IPv6 address
- [ ] Test external IPv6 connectivity: `ping6 google.com`

### 1.2 LAN IPv6 (Router Advertisements)
- [ ] Configure LAN interface (`igb0`) with IPv6
  - Use a `/64` from the delegated prefix for LAN
  - Enable Router Advertisements (RA) via SLAAC
  - Consider: SLAAC only vs SLAAC + DHCPv6 for address assignment
- [ ] Set RA mode to "Unmanaged" (SLAAC) initially for simplicity
- [ ] Enable "Advertise Default Gateway" in RA settings

### 1.3 IPv6 Firewall Rules
- [ ] Create WAN IPv6 rules:
  - Allow ICMPv6 (required for NDP, path MTU discovery)
  - Block all inbound by default (same as IPv4)
  - Allow established/related traffic
- [ ] Create LAN IPv6 rules:
  - Allow all outbound (match IPv4 policy)
  - Allow ICMPv6 between LAN hosts
- [ ] **Critical:** Do NOT block ICMPv6 types 133-137 (NDP) — this breaks IPv6

### 1.4 DNS64/NAT64 (Optional — Skip Initially)
- Not needed for dual-stack; only for IPv6-only networks

---

## Phase 2: Pi-hole DNS Configuration

### 2.1 Enable IPv6 Listening
- [ ] Edit Pi-hole config to listen on IPv6 interface
- [ ] Verify Pi-hole gets an IPv6 address via SLAAC
- [ ] Test: `dig AAAA google.com @<pihole-ipv6-address>`

### 2.2 Upstream DNS over IPv6
- [ ] Add IPv6 upstream resolvers:
  - Quad9: `2620:fe::11`
  - Cloudflare: `2606:4700:4700::1111`
  - Google: `2001:4860:4860::8888`
- [ ] Verify Pi-hole can resolve over IPv6

### 2.3 Advertise Pi-hole as IPv6 DNS
- [ ] In OPNsense RA settings, advertise Pi-hole's IPv6 address as DNS server
- [ ] Or use DHCPv6 to push DNS server (if using managed mode)

---

## Phase 3: Server/VM IPv6 Configuration

### 3.1 Verify SLAAC on Each Host
For each host, verify it receives an IPv6 address via SLAAC:

| Host | Expected Result | Command |
|------|----------------|---------|
| sentinella.alpina | Gets `2603:8001:...` address | `ip -6 addr show` |
| ntp.alpina | Gets `2603:8001:...` address | `ip -6 addr show` |
| komga.alpina | Gets `2603:8001:...` address | `ip -6 addr show` |
| aria.alpina (Proxmox) | Gets `2603:8001:...` address | `ip -6 addr show` |
| gotra | Gets `2603:8001:...` address | `ip -6 addr show` |
| pihole | Gets `2603:8001:...` address | `ip -6 addr show` |

### 3.2 Accept Router Advertisements
- [ ] Ensure `accept_ra = 1` on all Linux hosts:
  ```bash
  sysctl net.ipv6.conf.all.accept_ra
  # Should be 1 (default on most distros)
  ```
- [ ] For Proxmox (which runs as a router), may need `accept_ra = 2`

### 3.3 Service Binding
Verify services bind to IPv6 as well:
- [ ] node_exporter: Already binds to `:::9100` (all interfaces including IPv6)
- [ ] Chrony (NTP): Add `bindaddress ::` if needed
- [ ] Komga landing page: Go binary likely already dual-stack
- [ ] Sentinella stack: Podman containers — verify port bindings include IPv6

---

## Phase 4: DNS Records (Pi-hole Custom DNS)

### 4.1 Add AAAA Records
Once hosts have stable IPv6 addresses, add AAAA records to Pi-hole:

```
# /etc/pihole/custom.list (add AAAA records)
# Format: <ipv6-address> <hostname>
# These will be populated once SLAAC addresses are assigned
```

**Note:** SLAAC addresses are derived from MAC addresses (EUI-64) or random (privacy extensions). For stable DNS records, consider:
- Using static IPv6 addresses for servers
- Or disabling privacy extensions on servers: `sysctl net.ipv6.conf.all.use_tempaddr=0`

### 4.2 Stable Addresses for Servers
For predictable server addresses, use static IPv6 from the delegated prefix:

| Host | Suggested IPv6 (within /64) |
|------|----------------------------|
| gateway.alpina | `<prefix>::1` |
| pihole | `<prefix>::66` |
| ntp.alpina | `<prefix>::108` |
| komga.alpina | `<prefix>::202` |
| sentinella.alpina | `<prefix>::94` |
| aria.alpina | `<prefix>::230` |

(Last octets match IPv4 scheme where practical)

---

## Phase 5: Monitoring & Observability

### 5.1 Prometheus Scraping over IPv6
- [ ] Add IPv6 targets to prometheus.yml (or use hostnames that resolve to AAAA)
- [ ] Verify node_exporter accessible over IPv6

### 5.2 Grafana Dashboard Updates
- [ ] Add IPv6 connectivity indicator to Command Center
- [ ] Add panel showing IPv6 traffic vs IPv4 traffic
- [ ] Monitor ICMPv6 RA failures (currently seeing errors — should resolve)

### 5.3 Syslog over IPv6
- [ ] Update rsyslog configs to also send over IPv6 if desired
- [ ] Alloy syslog receiver already listens on `:::1514`

---

## Phase 6: Testing & Validation

### 6.1 Connectivity Tests
```bash
# From each host:
ping6 google.com                    # External IPv6
ping6 <gateway-ipv6-ll>            # Gateway reachability
ping6 <pihole-ipv6>                # DNS server
curl -6 https://ipv6.google.com    # HTTPS over IPv6
```

### 6.2 DNS Resolution
```bash
dig AAAA google.com @<pihole-ipv6>  # External resolution
dig AAAA ntp.alpina @<pihole-ipv6>  # Internal resolution
```

### 6.3 Service Access
```bash
curl -6 https://grafana.sentinella.alpina  # Grafana over IPv6
curl -6 http://ntp.alpina:8080              # NTP landing page
```

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| ISP changes delegated prefix | All IPv6 addresses change | Use hostnames, not hard-coded IPs |
| Broken ICMPv6 filtering | IPv6 stops working entirely | Always allow ICMPv6 types 1-4, 128-137 |
| Privacy extensions change addresses | DNS records stale | Use static IPs for servers |
| Application only binds to IPv4 | Service unreachable on IPv6 | Test each service; update bind address |
| IPv6 bypasses Pi-hole ad blocking | Ads return | Ensure Pi-hole is the IPv6 DNS server |

---

## Rollback Plan

If IPv6 causes issues:
1. Disable RA on OPNsense LAN interface (stops IPv6 address assignment)
2. Remove IPv6 firewall rules
3. Remove AAAA records from Pi-hole
4. Hosts will fall back to IPv4 automatically (dual-stack)

**IPv6 on the WAN side can remain enabled** even if LAN IPv6 is disabled.

---

## Order of Operations (Summary)

1. **OPNsense WAN** — Enable DHCPv6-PD, verify external connectivity
2. **OPNsense Firewall** — Add IPv6 rules (especially ICMPv6)
3. **OPNsense LAN** — Enable RA with SLAAC
4. **Pi-hole** — Enable IPv6 listening, add upstream resolvers
5. **Servers** — Verify SLAAC, assign static addresses, update DNS
6. **Monitoring** — Add IPv6 panels, verify scraping
7. **Test** — End-to-end connectivity and service access

---

## References

- [OPNsense IPv6 Configuration](https://docs.opnsense.org/manual/ipv6.html)
- [Pi-hole IPv6 Setup](https://docs.pi-hole.net/)
- [Charter/Spectrum IPv6 Info](https://www.spectrum.net/support/internet/ipv6)
- [RFC 4861 — Neighbor Discovery for IPv6](https://www.rfc-editor.org/rfc/rfc4861)
- [RFC 4862 — IPv6 SLAAC](https://www.rfc-editor.org/rfc/rfc4862)
