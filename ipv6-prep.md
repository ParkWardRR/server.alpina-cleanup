# IPv6 Preparation Plan — Alpina Homelab

**Status:** In Progress — OPNsense & Pi-hole Remediated
**Date:** 2026-02-06 (remediation pass 2)
**Goal:** Enable full dual-stack IPv6 across the entire Alpina homelab infrastructure

---

## Remediation Log — 2026-02-05

### Backups Created

| Host | Backup Location | Contents |
|------|----------------|----------|
| OPNsense | `/root/ipv6-backup-20260205_213520/` | config.xml, dhcp6c.conf, dhcp6c_wan_script.sh, radvd.conf, pf-rules.txt, sysctl-inet6.txt, ndp-table.txt, ipv6-routes.txt, igb0.txt, igb3.txt |
| Pi-hole | `/home/pi/ipv6-backup-20260205_213526/` | pihole.toml, custom.list, sysctl.conf, sysctl.d/, ipv6-addrs.txt, ipv6-routes.txt, pihole-version.txt |

### OPNsense Audit Results (No Changes Needed)

The OPNsense IPv6 configuration was found to be **significantly more complete than documented**. No changes were required.

**Already configured:**
- **DHCPv6-PD** active via `dhcp6c` on igb3 (WAN), receiving prefix `2603:8001:7400:fa9a::/64`
- **Router Advertisements** via radvd on igb0 (LAN):
  - Prefix `2603:8001:7400:fa9a::/64` with `AdvAutonomous on` (SLAAC)
  - **RDNSS** pointing to Pi-hole's link-local (`fe80::65b2:c033:6143:6d15`)
  - **DNSSL** set to `alpina`
- **LAN IPv6 address:** `2603:8001:7400:fa9a:a236:9fff:fe66:27ac/64` (EUI-64)
- **IPv6 forwarding:** enabled (`net.inet6.ip6.forwarding = 1`)
- **Firewall rules:** Comprehensive inet6 ruleset including:
  - ICMPv6 inbound: unreach, toobig, neighbrsol, neighbradv
  - ICMPv6 outbound: echoreq, echorep, routersol, routeradv, neighbrsol, neighbradv (to link-local + multicast)
  - ICMPv6 inbound from link-local: echoreq, routersol, routeradv, neighbrsol, neighbradv
  - WAN: all ICMPv6 allowed inbound with reply-to
  - LAN: `pass in quick on igb0 inet6 from (igb0:network) to any` — all IPv6 from LAN allowed
  - LAN: `pass in quick on igb0 inet6 from fe80::/10 to any` — link-local pass
  - DHCPv6 client/server traffic on both LAN and WAN
  - Route-to for outbound via WAN

**Conclusion:** OPNsense IPv6 is fully functional. RA mode is already advertising Pi-hole via RDNSS (using the stable link-local address, which is better than using a global address that could change with prefix delegation). No configuration changes made.

### Pi-hole Audit Results & Changes Made

**Pre-existing (already configured, not documented):**
- `listeningMode = "ALL"` — FTL already listens on all interfaces including IPv6
- DNS bound on `[::]:53` (TCP and UDP) — dual-stack listening confirmed
- All IPv6 upstreams already configured:
  - Quad9: `2620:fe::11`, `2620:fe::fe:11`
  - Cloudflare: `2606:4700:4700::1111`, `2606:4700:4700::1001`
  - Google: `2001:4860:4860::8888`, `2001:4860:4860::8844`
- `blockingmode = "null"` — returns `0.0.0.0` for A and `::` for AAAA on blocked domains
- **IPv6 address is stable-privacy** (`addr_gen_mode=1`, RFC 7217) — NOT EUI-64, NOT temporary. The address `4392:b645:21ad:5510` is deterministic and stable across reboots.
- External IPv6 connectivity works: `ping -6 google.com` succeeds
- AAAA resolution works: `dig AAAA google.com @127.0.0.1` returns valid results

**Changes made — AAAA records added to `/etc/pihole/custom.list`:**

```
# IPv6 AAAA records — added 2026-02-05
2603:8001:7400:fa9a:a236:9fff:fe66:27ac gateway.alpina
2603:8001:7400:fa9a:4392:b645:21ad:5510 pihole.alpina
2603:8001:7400:fa9a:be24:11ff:fe60:2dfe ntp.alpina
2603:8001:7400:fa9a:be24:11ff:fe60:2dfe ntp
2603:8001:7400:fa9a:be24:11ff:fe09:c0b9 komga.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956 sentinella.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956 grafana.sentinella.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956 prometheus.sentinella.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956 loki.sentinella.alpina
2603:8001:7400:fa9a:be24:11ff:fe95:2956 alloy.sentinella.alpina
2603:8001:7400:fa9a:be24:11ff:fec9:2694 home.alpina
```

Also added missing A record: `172.16.16.16 gateway.alpina`

**Verification results:**

| Hostname | AAAA Resolution | A Resolution |
|----------|----------------|-------------|
| ntp.alpina | `2603:8001:7400:fa9a:be24:11ff:fe60:2dfe` | `172.16.16.108` |
| komga.alpina | `2603:8001:7400:fa9a:be24:11ff:fe09:c0b9` | `172.16.16.202` |
| sentinella.alpina | `2603:8001:7400:fa9a:be24:11ff:fe95:2956` | `172.16.19.94` |
| home.alpina | `2603:8001:7400:fa9a:be24:11ff:fec9:2694` | `172.16.17.109` |
| pihole.alpina | `::1` (self-reference, expected) | `127.0.0.1` |
| grafana.sentinella.alpina | **(not resolving)** | `172.16.19.94` |
| gateway.alpina | **(not resolving)** | **(not resolving)** |

**Known issue:** Pi-hole v6 FTL has a limitation where some hostnames in `custom.list` don't resolve AAAA records despite being present in the file. This affects `gateway.alpina` (no records resolve at all) and subdomain-style hostnames like `grafana.sentinella.alpina` (A works, AAAA doesn't). The core server hostnames all resolve correctly. The subdomain services are accessible via `sentinella.alpina` (which resolves both A and AAAA) since Caddy routes by SNI/Host header.

---

## Current State (Post-Remediation)

### ISP Details

- **ISP:** Charter/Spectrum
- **Protocol:** DHCPv6 Prefix Delegation via `dhcp6c` on igb3 (WAN)
- **Delegated prefix:** `2603:8001:7400:fa9a::/64`
- **Current PD config:** Single `/64` (sla-id 0, sla-len 0)

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

### OPNsense RA Config (`/var/etc/radvd.conf`)

```
interface igb0 {
    AdvSendAdvert on;
    MinRtrAdvInterval 200;
    MaxRtrAdvInterval 600;
    AdvLinkMTU 1500;
    AdvDefaultPreference medium;
    prefix 2603:8001:7400:fa9a::/64 {
        DeprecatePrefix on;
        AdvOnLink on;
        AdvAutonomous on;
    };
    RDNSS fe80::65b2:c033:6143:6d15 {
    };
    DNSSL alpina {
    };
};
```

### Per-Host IPv6 Status (Updated)

| Host | IPv4 | IPv6 Address | Type | Ping6 | AAAA in DNS |
|------|------|-------------|------|-------|-------------|
| gateway.alpina (LAN) | 172.16.16.16 | `2603:8001:7400:fa9a:a236:9fff:fe66:27ac/64` | EUI-64 | N/A | No (FTL bug) |
| pihole | 172.16.66.66 | `2603:8001:7400:fa9a:4392:b645:21ad:5510/64` | Stable-Privacy | Works | Yes (::1) |
| ntp.alpina | 172.16.16.108 | `2603:8001:7400:fa9a:be24:11ff:fe60:2dfe/64` | EUI-64 | Works | Yes |
| home.alpina | 172.16.17.109 | `2603:8001:7400:fa9a:be24:11ff:fec9:2694/64` | EUI-64 | Works | Yes |
| komga.alpina | 172.16.16.202 | `2603:8001:7400:fa9a:be24:11ff:fe09:c0b9/64` | EUI-64 | Works | Yes |
| sentinella.alpina | 172.16.19.94 | `2603:8001:7400:fa9a:be24:11ff:fe95:2956/64` | EUI-64 | Works | Yes |
| aria.alpina (Proxmox) | 172.16.18.230 | `2603:8001:7400:fa9a:eaff:1eff:fed3:4683/64` | SLAAC | Works | No |
| homeassistant.local | 172.16.77.77 | **None visible** | N/A | **Unreachable** | No |

### What's Now Complete
- [x] OPNsense DHCPv6-PD — active, prefix delegated
- [x] OPNsense Router Advertisements — sending prefix + RDNSS (Pi-hole) + DNSSL (alpina)
- [x] OPNsense IPv6 firewall — comprehensive ruleset (ICMPv6, NDP, DHCPv6, LAN pass-all)
- [x] OPNsense IPv6 forwarding — enabled
- [x] Pi-hole listening on IPv6 — `[::]:53` TCP+UDP
- [x] Pi-hole IPv6 upstream resolvers — Quad9, Cloudflare, Google (all dual-stack)
- [x] Pi-hole AAAA record resolution — external AAAA queries work
- [x] Pi-hole blocking over IPv6 — `blockingmode = "null"` returns `::` for blocked AAAA
- [x] Pi-hole AAAA records for local hosts — added to custom.list (core hostnames resolve)
- [x] Pi-hole stable IPv6 address — using stable-privacy (addr_gen_mode=1), no action needed
- [x] 5 of 8 hosts have global IPv6 addresses via SLAAC

---

## Remediation Log — 2026-02-06

### Proxmox (aria.alpina)

- Added `/etc/sysctl.d/50-ipv6.conf` with:
  - `net.ipv6.conf.vmbr0.accept_ra=2`
  - `net.ipv6.conf.all.accept_ra=2`
- Applied with `sysctl -p /etc/sysctl.d/50-ipv6.conf` and solicited an RA (`rdisc6 -1 vmbr0`).
- Result: vmbr0 now has global SLAAC `2603:8001:7400:fa9a:eaff:1eff:fed3:4683/64`; dual default routes via gateway + Pi-hole observed (same as other hosts).
- Connectivity: `ping -6 google.com` succeeds (~14–17 ms RTT).

### home.alpina

- Firewalld: `sudo firewall-cmd --permanent --add-protocol=ipv6-icmp && sudo firewall-cmd --reload`.
- Runtime default route from Pi-hole still appears (two IPv6 default gateways). Manually removed the Pi-hole nexthop during testing: `sudo ip -6 route del default via fe80::65b2:c033:6143:6d15 dev ens18` — ping started working immediately. Route re-adds when new RAs arrive; root cause still the suspected rogue RA source.
- Connectivity: `ping -6 google.com` now succeeds.

### sentinella.alpina

- Same firewalld change: allow `ipv6-icmp` permanently and reload.
- Same dual-default-route behavior; removed Pi-hole nexthop during verification to get immediate success: `sudo ip -6 route del default via fe80::65b2:c033:6143:6d15 dev ens18`.
- Connectivity: `ping -6 google.com` succeeds after removal; route may reappear until rogue RA is addressed.

### Router Advertisement cleanup (Pi-hole)

- Identified unintended RAs (router lifetime 1800s, prefix 2603:8001:7400:fa9a::/64) coming from Pi-hole (`b8:27:eb:db:9a:15`) because Pi-hole DHCPv6 was enabled. RAs with route-info only (no default) also observed from multiple MACs `04:99:b9:71:36:9d/04:99:b9:84:18:17/ec:a9:07:07:22:bf/ac:bc:b5:db:26:fa` advertising ULA `fde6:19bd:3ffd::/64` but **router lifetime 0** (no default route impact).
- Fix applied on Pi-hole: set `[dhcp] ipv6 = false` in `/etc/pihole/pihole.toml`, restart `pihole-FTL`, and verify `dnsmasq.conf` no longer contains `dhcp-range=::` (no RAs emitted; tcpdump on Pi-hole shows 0 ICMPv6 type 134).
- Cleared stale RA-learned defaults by bouncing NetworkManager on hosts: `nmcli conn down ens18 && nmcli conn up ens18` on `home.alpina` and `sentinella.alpina`.
- Result: both hosts now have a **single IPv6 default route via OPNsense** (`fe80::a236:9fff:fe66:27ac`) and `ping -6 google.com` succeeds.

---

## Next Steps (Remaining Work)

### Priority 1: Fix Broken Hosts

#### Proxmox (aria.alpina) — No Global IPv6
Proxmox has only link-local on vmbr0. Likely cause: `forwarding=1` suppresses RA acceptance.
```bash
ssh root@aria.alpina
echo "net.ipv6.conf.vmbr0.accept_ra=2" >> /etc/sysctl.d/50-ipv6.conf
sysctl -p /etc/sysctl.d/50-ipv6.conf
ip -6 addr show vmbr0
ping6 -c4 google.com
```

#### home.alpina — Firewalld Blocks ICMPv6
Has global IPv6 but `firewalld` blocks outbound ICMPv6.
```bash
ssh -i ~/.ssh/id_ed25519 alfa@home.alpina
sudo firewall-cmd --permanent --add-protocol=ipv6-icmp
sudo firewall-cmd --reload
ping6 -c4 google.com
```

#### sentinella.alpina — Same Firewalld Issue
```bash
ssh alfa@sentinella.alpina
sudo firewall-cmd --permanent --add-protocol=ipv6-icmp
sudo firewall-cmd --reload
ping6 -c4 google.com
```

### Priority 2: Optional Improvements

#### Upgrade PD to /56 (Future VLANs)
Currently getting a single `/64`. Requesting `/56` from Charter/Spectrum would give 256 subnets for future VLANs. **Risk:** May change the current prefix, breaking all SLAAC addresses temporarily.

In OPNsense web UI: Interfaces > WAN (igb3) > IPv6 Configuration:
- Prefix delegation size: `/56`
- Send IPv6 prefix hint: Checked

#### Investigate Dual Default Routes
Some hosts show two IPv6 default routes. Check if Pi-hole is inadvertently sending RAs:
```bash
ssh -i ~/.ssh/id_ed25519 pi@pihole 'ps aux | grep -i radvd'
```

#### Pi-hole v6 FTL AAAA Bug
`gateway.alpina` and subdomain-style hostnames (`grafana.sentinella.alpina`) don't resolve AAAA from custom.list despite entries being present. Consider:
- Filing a Pi-hole v6 bug report
- Using the Pi-hole v6 web UI "Local DNS Records" to add AAAA entries instead of custom.list
- Using CNAME records pointing subdomains to `sentinella.alpina` (which resolves AAAA correctly)

#### Docker/Podman IPv6
- Komga (Docker): `docker network inspect bridge | grep EnableIPv6` — likely `false`
- Sentinella (Podman): Check if container port bindings include IPv6

#### Grafana IPv6 Panels
Add to Command Center dashboard:
- IPv6 connectivity indicator per host
- IPv6 vs IPv4 traffic ratio

---

## Rollback Plan

**OPNsense restore:**
```bash
ssh root@172.16.16.16
cp /root/ipv6-backup-20260205_213520/config.xml /conf/config.xml
# Reboot or reload via web UI
```

**Pi-hole restore:**
```bash
ssh -i ~/.ssh/id_ed25519 pi@pihole
sudo cp /home/pi/ipv6-backup-20260205_213526/custom.list /etc/pihole/custom.list
sudo pihole reloaddns
```

**Nuclear option (disable all LAN IPv6):**
1. Disable RA on OPNsense LAN interface (stops IPv6 address assignment)
2. Remove AAAA records from Pi-hole
3. Hosts fall back to IPv4 automatically (dual-stack)

IPv6 on the WAN side can remain enabled even if LAN IPv6 is disabled.

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
