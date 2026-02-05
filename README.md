# Alpina Homelab

Configuration and documentation for the Alpina homelab network.

## Quick Reference

| Service | Host | IP |
|---------|------|-----|
| Firewall | gateway.alpina | 172.16.16.16 |
| DNS/Ad-block | pihole | 172.16.66.66 |
| NTP | ntp.alpina | 172.16.16.108 |
| Komga | komga.alpina | 172.16.16.202 |

## SSH Access

```bash
ssh root@172.16.16.16          # OPNsense
ssh pi@pihole                   # Pi-hole
ssh alfa@ntp.alpina             # NTP server
```

## Documentation

See [homelab-details.md](homelab-details.md) for full technical details including network topology, service configurations, and custom DNS records.
