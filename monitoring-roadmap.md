# Monitoring Roadmap

## Vision
A comprehensive observability platform providing real-time insights into the Alpina homelab infrastructure, enabling proactive monitoring, quick troubleshooting, and security awareness.

---

## Current State (v1.0)

### Deployed
- [x] Sentinella observability server (AlmaLinux 10.1)
- [x] Grafana dashboards with Loki datasource
- [x] Prometheus metrics collection
- [x] Alloy syslog receiver (UDP 1514)
- [x] Caddy reverse proxy with TLS
- [x] Landing page portal

### Log Sources Active
- [x] Proxmox (aria.alpina)
- [x] Pi-hole DNS server
- [x] OPNsense firewall
- [x] NTP server (chrony)
- [x] Gotra application server

### Metrics Sources Active
- [x] Node exporter on all Linux hosts (5 hosts)
- [x] Prometheus self-monitoring
- [x] Grafana metrics
- [x] Loki metrics
- [x] Alloy metrics

---

## Short Term (v1.1) - COMPLETED ✅

### Log Source Expansion
- [x] OPNsense firewall logs
- [x] NTP server (chrony) logs
- [x] Gotra application server
  - [x] System logs via rsyslog
  - [x] Gunicorn access/error logs
  - [x] Celery worker logs

### Dashboard Improvements
- [x] NTP Time Sync Dashboard
- [x] Firewall Security Dashboard (OPNsense)
- [x] Application Performance Dashboard (Gotra)
- [x] Home page with health overview
- [x] System Metrics Dashboard

### Data Retention
- [x] Configured 24-month retention for Prometheus
- [x] Configured 24-month retention for Loki

### Structured Log Parsing (Alloy)
- [x] OPNsense filterlog parsing
- [x] SSH authentication parsing
- [x] NTP chrony parsing
- [x] Celery task parsing
- [x] Pi-hole DNS parsing

### Alerting
- [ ] Grafana alerting rules
- [ ] Email/webhook notifications
- [ ] Critical event detection

---

## Medium Term (v1.2)

### Additional Log Sources
- [ ] Komga media server
- [ ] Docker containers on various hosts
- [ ] Network switches (if syslog capable)
- [ ] UPS status logs

### Metrics Expansion
- [x] Node exporter on all Linux hosts
- [ ] cAdvisor for container metrics
- [ ] Custom application metrics
- [ ] Network bandwidth monitoring

### Security Monitoring
- [ ] Failed login attempt tracking
- [ ] Unusual network traffic detection
- [ ] SSH key usage auditing
- [ ] Certificate expiry monitoring

---

## Long Term (v2.0)

### Infrastructure
- [ ] High availability setup
- [ ] Long-term storage (S3/MinIO for Loki)
- [ ] Backup and disaster recovery
- [ ] Grafana OnCall integration

### Advanced Features
- [ ] Log-based anomaly detection
- [ ] Automated incident response
- [ ] Capacity planning dashboards
- [ ] Cost tracking (if applicable)

### Integrations
- [ ] Home Assistant events
- [ ] GitHub Actions webhooks
- [ ] External uptime monitoring
- [ ] PagerDuty/Slack alerts

---

## Technical Debt

### High Priority
- [x] Enable Loki retention policy (24 months)
- [x] Enable Prometheus retention policy (24 months)
- [ ] Add health checks to all containers
- [ ] Implement log rotation on source systems
- [ ] Document backup procedures

### Medium Priority
- [ ] Optimize Prometheus scrape intervals
- [ ] Add Grafana plugins for better visualization
- [ ] Create runbooks for common issues
- [ ] Performance tuning for high-volume logs

### Low Priority
- [ ] Migrate to newer container images (when stable)
- [ ] Explore Tempo for distributed tracing
- [ ] Consider Mimir for long-term metrics

---

## Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Log sources | 5+ | 5 (Proxmox, Pi-hole, OPNsense, NTP, Gotra) |
| Metrics sources | 5+ | 5 hosts with node_exporter |
| Dashboard load time | <3 seconds | ~2 seconds |
| Log retention | 24 months | 24 months ✅ |
| Metrics retention | 24 months | 24 months ✅ |
| System uptime | 99.9% | TBD |

---

## Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Alloy Documentation](https://grafana.com/docs/alloy/)

---

## Contributing

To add a new log source:
1. Configure syslog forwarding to `sentinella.alpina:1514/udp`
2. Verify logs appear in Loki
3. Create or update Grafana dashboard
4. Update this roadmap and history documents
5. Commit changes to main branch
