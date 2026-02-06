# IPv6 — Current State & History (as of 2026-02-06)

## Summary
- Prefix: `2603:8001:7400:fa9a::/64` via DHCPv6-PD (Spectrum) on OPNsense.
- OPNsense: RAs on LAN with RDNSS (Pi-hole link-local), DNSSL=alpina; IPv6 firewall open for ICMPv6/NDP/DHCPv6; forwarding enabled.
- Pi-hole: Dual-stack DNS; DHCPv6/RA **disabled** to stop rogue RAs; upstream IPv6 resolvers configured; custom AAAA records for core hosts.
- Working dual-stack hosts: gateway, pihole, ntp, komga, home, sentinella, aria (vmbr0), portocali — all reachable via IPv6 and internet `ping -6` succeeds.
- Outstanding: HAOS has no global IPv6; route-info-only RAs for ULA `fde6:19bd:3ffd::/64` from MACs `04:99:b9:71:36:9d`, `04:99:b9:84:18:17`, `ec:a9:07:07:22:bf`, `ac:bc:b5:db:26:fa` (identify and disable source).

## Timeline
- 2026-02-05: Audited OPNsense (already fully configured); Pi-hole dual-stack verified; added AAAA records; 5/8 hosts had IPv6; home/sentinella timed out; Proxmox link-local only.
- 2026-02-06 (Pass 2):
  - Proxmox: Added `/etc/sysctl.d/50-ipv6.conf` with `accept_ra=2` (vmbr0/all); now global SLAAC `2603:...:fed3:4683` and internet IPv6 works.
  - home.alpina & sentinella.alpina: firewalld now allows `ipv6-icmp`; IPv6 connectivity restored.
  - Pi-hole: Disabled DHCPv6/RA (`[dhcp] ipv6=false` in `pihole.toml`, restart FTL); stopped default-route RAs from Pi-hole; bounced NM on hosts to clear stale defaults → single default via gateway.
  - Portocali NAS: IPv6 enabled in DSM network settings; SLAAC address `2603:...:7656:3cff:fe30:2dfc` (EUI-64); default route via gateway; 8/9 hosts now dual-stack.

## Per-Host Status

| Host | IPv4 | IPv6 | Type | Ping6 | Notes |
|------|------|------|------|-------|-------|
| gateway.alpina | 172.16.16.16 | 2603:8001:7400:fa9a:a236:9fff:fe66:27ac/64 | EUI-64 | n/a | RA/RDNSS/DNSSL source |
| pihole | 172.16.66.66 | 2603:8001:7400:fa9a:4392:b645:21ad:5510/64 | stable-privacy | Works | DHCPv6 off; dual-stack DNS |
| ntp.alpina | 172.16.16.108 | 2603:8001:7400:fa9a:be24:11ff:fe60:2dfe/64 | EUI-64 | Works | — |
| komga.alpina | 172.16.16.202 | 2603:8001:7400:fa9a:be24:11ff:fe09:c0b9/64 | EUI-64 | Works | — |
| home.alpina | 172.16.17.109 | 2603:8001:7400:fa9a:be24:11ff:fec9:2694/64 | EUI-64 | Works | firewalld allows ipv6-icmp |
| sentinella.alpina | 172.16.19.94 | 2603:8001:7400:fa9a:be24:11ff:fe95:2956/64 | EUI-64 | Works | firewalld allows ipv6-icmp |
| aria.alpina (Proxmox) | 172.16.18.230 | 2603:8001:7400:fa9a:eaff:1eff:fed3:4683/64 | SLAAC | Works | accept_ra=2 on vmbr0/all |
| portocali.alpina | 172.16.21.21 | 2603:8001:7400:fa9a:7656:3cff:fe30:2dfc/64 | EUI-64 | Works | Enabled in DSM |
| homeassistant.alpina | 172.16.77.77 | — | — | Unreachable | HAOS: no global IPv6 |

## Actions Remaining
- Identify & disable ULA `fde6:19bd:3ffd::/64` RA sources (MACs above).
- Investigate HAOS IPv6 support/limitations; enable if possible.
- Add AAAA record for `portocali.alpina` in Pi-hole custom.list.
- Optional: request /56 PD from Spectrum (may renumber prefix).

## Backups
- OPNsense: `/root/ipv6-backup-20260205_213520/` (config, radvd, dhcp6c, pf rules, ndp, routes).
- Pi-hole: `/home/pi/ipv6-backup-20260205_213526/` (pihole.toml, custom.list, sysctl*, route snapshots).

