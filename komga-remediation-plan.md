# Komga Server Remediation Plan

**Server:** komga.alpina (172.16.16.202)
**OS:** Debian 12 (bookworm)
**Hardware:** Proxmox VM - 4 vCPU, 4.2GB RAM, 392GB disk
**Assessment Date:** 2026-02-04

---

## Executive Summary

| Category | Status | Priority |
|----------|--------|----------|
| Komga Service | ❌ Down (8 months) | CRITICAL |
| NFS Mount | ❌ Missing | CRITICAL |
| Firewall | ❌ None | HIGH |
| SSH Hardening | ⚠️ Partial | HIGH |
| Auto Updates | ❌ Not configured | HIGH |
| Fail2ban | ❌ Not installed | MEDIUM |
| Tailscale | ⚠️ Logged out | LOW |
| Time Sync | ✅ Working | - |

---

## Critical Issues

### 1. Komga Container Down (8 months)

**Problem:** Container exited with code 128 on 2025-06-04. Data mount `/mnt/MonterosaSync-Read` is empty/unmounted.

**Fix:**
```bash
# 1. Identify the NFS share source (check your NAS/file server)
# Example: If it's from a Synology/TrueNAS at 172.16.x.x

# 2. Install NFS client (already have rpcbind)
sudo apt install nfs-common

# 3. Test mount manually
sudo mount -t nfs <NAS_IP>:/path/to/share /mnt/MonterosaSync-Read

# 4. Add to /etc/fstab for persistence
echo "<NAS_IP>:/path/to/share /mnt/MonterosaSync-Read nfs defaults,_netdev,nofail 0 0" | sudo tee -a /etc/fstab

# 5. Verify and start Komga
sudo mount -a
docker start komga
docker logs -f komga
```

### 2. No Firewall Configured

**Problem:** No ufw or iptables rules. Server is exposed on LAN.

**Fix:**
```bash
# Install and configure UFW
sudo apt install ufw

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow necessary services
sudo ufw allow ssh
sudo ufw allow from 172.16.0.0/16 to any port 25600 proto tcp comment 'Komga web UI'

# Enable firewall
sudo ufw enable
sudo ufw status verbose
```

---

## High Priority

### 3. SSH Hardening

**Problem:** Default SSH config, password auth may be enabled.

**Fix:**
```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config.d/hardening.conf
```

Add:
```
PasswordAuthentication no
PermitRootLogin no
X11Forwarding no
MaxAuthTries 3
ClientAliveInterval 300
ClientAliveCountMax 2
```

```bash
# Restart SSH
sudo systemctl restart sshd
```

### 4. Automatic Security Updates

**Problem:** No unattended-upgrades, package lists are 1 year old.

**Fix:**
```bash
# Update system first
sudo apt update && sudo apt upgrade -y

# Install unattended-upgrades
sudo apt install unattended-upgrades apt-listchanges

# Configure for security updates only
sudo dpkg-reconfigure -plow unattended-upgrades

# Verify
sudo unattended-upgrades --dry-run
```

### 5. Install Fail2ban

**Problem:** No brute-force protection.

**Fix:**
```bash
sudo apt install fail2ban

# Create local config
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo nano /etc/fail2ban/jail.local
```

Add/modify:
```ini
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
findtime = 600
```

```bash
sudo systemctl enable --now fail2ban
sudo fail2ban-client status sshd
```

---

## Medium Priority

### 6. Update Komga Container

**Problem:** Komga 1.19.1 is 12 months old (current is ~1.20+).

**Fix:**
```bash
# Pull latest image
docker pull ghcr.io/gotson/komga:latest

# Stop old container
docker stop komga

# Remove old container (data is in volumes)
docker rm komga

# Start new container with same config
docker run -d \
  --name komga \
  --restart unless-stopped \
  -p 25600:25600 \
  -v /home/alfa/komga/config:/config \
  -v /home/alfa/komga/logs:/logs \
  -v /mnt/MonterosaSync-Read:/data:ro \
  -e TZ=America/Los_Angeles \
  ghcr.io/gotson/komga:latest
```

### 7. Disable Unused Services

**Problem:** RPC services running but no NFS exports.

**Fix:**
```bash
# If not using NFS server features
sudo systemctl disable --now rpcbind
sudo systemctl disable --now rpc-statd

# Note: Re-enable if needed for NFS client
```

### 8. Configure NTP to Use Local Server

**Problem:** Using systemd-timesyncd instead of local NTP server.

**Fix:**
```bash
# Edit timesyncd config
sudo nano /etc/systemd/timesyncd.conf
```

Add:
```ini
[Time]
NTP=172.16.16.108
FallbackNTP=time.google.com
```

```bash
sudo systemctl restart systemd-timesyncd
timedatectl timesync-status
```

---

## Low Priority

### 9. Reconnect Tailscale (Optional)

**Problem:** Tailscale is logged out.

**Fix:**
```bash
sudo tailscale up
# Follow the login URL provided
```

### 10. Clean Up Docker

**Fix:**
```bash
# Remove unused images/volumes
docker system prune -a --volumes
```

---

## Implementation Order

1. **Fix NFS mount** (required for Komga)
2. **Start Komga container** (restore service)
3. **Update system packages** (security)
4. **Install unattended-upgrades** (ongoing security)
5. **Configure UFW firewall** (network security)
6. **Harden SSH** (access security)
7. **Install fail2ban** (brute-force protection)
8. **Update Komga** (application security)
9. **Configure local NTP** (time accuracy)
10. **Clean up Docker** (maintenance)

---

## Verification Commands

```bash
# Check Komga
docker ps | grep komga
curl -I http://localhost:25600

# Check firewall
sudo ufw status numbered

# Check SSH hardening
sudo sshd -T | grep -E 'password|permit|x11'

# Check fail2ban
sudo fail2ban-client status

# Check auto-updates
sudo systemctl status unattended-upgrades

# Check NFS mount
df -h /mnt/MonterosaSync-Read
```

---

## Questions Before Implementation

1. **What is the NFS share source?** (IP and path for `/mnt/MonterosaSync-Read`)
2. **Do you want Tailscale reconnected?**
3. **Any other services planned for this server?**

---

## Current Server Inventory

| Component | Value |
|-----------|-------|
| Hostname | komga.alpina |
| IP | 172.16.16.202 |
| OS | Debian 12 (bookworm) |
| Kernel | 6.1.0-30-amd64 |
| CPU | 4 vCPU (Intel N100) |
| RAM | 4.2 GB |
| Disk | 392 GB (3 GB used) |
| Docker | Installed |
| Komga | v1.19.1 (stopped) |
| Uptime | 28 minutes |
