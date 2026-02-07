# Tailscale Remote Access — Alpina Homelab

## Recommended Architecture

**OPNsense subnet router** — single install, full LAN access, zero changes to existing hosts.

```
                        ┌─────────────────────────┐
  Phone / Laptop        │   Tailscale Coordination │
  (Tailscale client) ◄──┤   (100.x.x.x mesh)      ├──► OPNsense (subnet router)
                        └─────────────────────────┘         │
                                                            │ advertises 172.16.0.0/12
                                                            ▼
                                                    ┌───────────────┐
                                                    │  Entire LAN   │
                                                    │  All 9 hosts  │
                                                    │  All services  │
                                                    └───────────────┘
```

Install Tailscale **only on OPNsense** as a subnet router advertising `172.16.0.0/12`. Every LAN service becomes reachable from any Tailscale-connected device — Grafana, Home Assistant, Komga, Proxmox, Pi-hole, NAS, SSH — without installing Tailscale on individual hosts.

### Why This Is the Best Approach

| Approach | Pros | Cons |
|----------|------|------|
| **Subnet router on OPNsense** | One install, all services accessible, no host changes, exit node capable | Single point of entry (OPNsense must be up) |
| Tailscale on every host | Direct mesh connections | 9 installs to maintain, unnecessary complexity |
| VPN server on OPNsense (WireGuard/OpenVPN) | No third-party dependency | Port forwarding, cert management, no SSO, manual key rotation |
| Cloudflare Tunnel | No open ports | Only HTTP/S services, no SSH/NTP/SNMP/SMB |

The subnet router wins because your OPNsense box is already the gateway — it has routes to everything and is always on. One install covers all 9 hosts and every protocol (HTTP, SSH, NFS, SMB, SNMP, syslog, NTP).

---

## Implementation

### Step 1: Install Tailscale on OPNsense

```bash
ssh root@172.16.16.16
```

Install the plugin:
```bash
pkg install os-tailscale
```

Or via OPNsense Web UI: **System > Firmware > Plugins** → search `os-tailscale` → install.

### Step 2: Authenticate and Configure Subnet Router

```bash
# Enable IP forwarding (should already be enabled on OPNsense)
sysctl net.inet.ip.forwarding=1
sysctl net.inet6.ip6.forwarding=1

# Start Tailscale as subnet router
tailscale up \
  --advertise-routes=172.16.0.0/12 \
  --accept-dns=false \
  --hostname=gateway-alpina
```

- `--advertise-routes=172.16.0.0/12` — Advertises the entire LAN range
- `--accept-dns=false` — Keeps Pi-hole as DNS on the gateway itself (don't override with Tailscale DNS)
- `--hostname=gateway-alpina` — Clean name in the Tailscale admin console

This prints a URL — open it to authenticate with your Tailscale account.

### Step 3: Approve Subnet Routes in Tailscale Admin

1. Go to https://login.tailscale.com/admin/machines
2. Find `gateway-alpina`
3. Click the **...** menu → **Edit route settings**
4. Enable the `172.16.0.0/12` subnet route

### Step 4: Configure Split DNS for `.alpina`

In the Tailscale admin console:

1. Go to **DNS** tab
2. Under **Nameservers** → **Add nameserver** → **Custom**
3. Add Pi-hole's Tailscale-reachable IP: `172.16.66.66`
4. **Restrict to domain**: `alpina`

This routes all `.alpina` DNS queries through Pi-hole, so `ntp.alpina`, `sentinella.alpina`, etc. resolve correctly from anywhere.

### Step 5: (Optional) Enable Exit Node

To route ALL traffic through your home network (not just LAN traffic):

```bash
tailscale up \
  --advertise-routes=172.16.0.0/12 \
  --advertise-exit-node \
  --accept-dns=false \
  --hostname=gateway-alpina
```

Then approve the exit node in the admin console. When traveling, enable exit node on your client to get Pi-hole ad blocking on all traffic.

---

## Access After Setup

From any device running the Tailscale client:

| Service | URL / Command | Notes |
|---------|---------------|-------|
| Grafana | https://grafana.sentinella.alpina | Full dashboard access |
| Home Assistant | http://homeassistant.alpina:8123 | Smart home control |
| Komga | http://komga.alpina:25600 | Comics/manga reader |
| Pi-hole | http://172.16.66.66/admin | DNS management |
| OPNsense | https://172.16.16.16 | Firewall management |
| Proxmox | https://aria.alpina:8006 | VM management |
| NAS (DSM) | http://portocali.alpina:5000 | File management |
| NTP Landing | http://ntp.alpina | NTP performance dashboard |
| SSH (any host) | `ssh alfa@ntp.alpina` | All hosts reachable |
| SMB Shares | `\\portocali.alpina\` | NAS file shares |

---

## ACL Policy (Tailscale Admin → Access Controls)

Restrict what devices can reach on your network:

```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["autogroup:owner"],
      "dst": ["*:*"]
    }
  ],
  "autoApprovers": {
    "routes": {
      "172.16.0.0/12": ["autogroup:owner"]
    }
  }
}
```

This gives full access to the account owner. To add shared access for other users later:

```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["autogroup:owner"],
      "dst": ["*:*"]
    },
    {
      "action": "accept",
      "src": ["group:guests"],
      "dst": [
        "172.16.77.77:8123",
        "172.16.16.202:25600"
      ]
    }
  ],
  "groups": {
    "group:guests": ["user@example.com"]
  }
}
```

This gives guests access to only Home Assistant and Komga.

---

## Optional: Backup Access via Sentinella

For resilience if OPNsense goes down, install Tailscale on Sentinella as a secondary entry point:

```bash
ssh alfa@sentinella.alpina
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --hostname=sentinella-alpina
```

This gives direct access to Grafana/Prometheus/Loki even if the gateway is unreachable. No subnet routing needed — just direct access to the monitoring stack.

---

## OPNsense Firewall Rules

No inbound port forwarding or firewall changes needed. Tailscale uses outbound connections (DERP relays or direct WireGuard via UDP hole-punching), so it works behind NAT with zero firewall changes.

If you want to ensure direct connections (lower latency), allow outbound UDP on port 41641 — but this is typically already allowed by default.

---

## Monitoring Tailscale

Add to Grafana Command Center:

- **Tailscale status:** `tailscale status --json` on OPNsense (via a simple script or node_exporter textfile collector)
- **Connection type:** Check if connections are direct (p2p) or relayed (DERP)
- **Subnet router health:** Monitor OPNsense uptime as it's the single entry point

---

## Tailscale vs Alternatives Summary

| Feature | Tailscale | WireGuard (manual) | OpenVPN | Cloudflare Tunnel |
|---------|-----------|-------------------|---------|-------------------|
| Setup time | ~10 min | ~1 hour | ~2 hours | ~30 min |
| Port forwarding | None needed | Yes (UDP) | Yes (UDP/TCP) | None needed |
| Protocol support | All (L3 tunnel) | All (L3 tunnel) | All (L3 tunnel) | HTTP/S only |
| SSO / MFA | Built-in (Google, Microsoft, etc.) | Manual keys | Cert-based | Cloudflare Access |
| ACLs | Web console | iptables | Server config | Cloudflare rules |
| NAT traversal | Automatic | Manual / STUN | Manual | N/A |
| Split DNS | Built-in | Manual | Manual | Partial |
| Cost (personal) | Free (100 devices) | Free | Free | Free (limited) |
| Mobile apps | iOS, Android | iOS, Android | iOS, Android | N/A |

---

## Quick Start Checklist

- [ ] Install `os-tailscale` on OPNsense
- [ ] Run `tailscale up --advertise-routes=172.16.0.0/12 --accept-dns=false --hostname=gateway-alpina`
- [ ] Approve subnet route in Tailscale admin console
- [ ] Add split DNS for `.alpina` → `172.16.66.66` (Pi-hole)
- [ ] Install Tailscale client on phone/laptop
- [ ] Test: `ssh alfa@ntp.alpina` from outside the LAN
- [ ] Test: Open http://homeassistant.alpina:8123 from phone on cellular
- [ ] (Optional) Enable exit node for full-tunnel + Pi-hole ad blocking
- [ ] (Optional) Install on sentinella.alpina as backup entry point
