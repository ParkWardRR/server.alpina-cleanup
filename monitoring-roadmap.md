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

---

## Short Term (v1.1) - In Progress

### Log Source Expansion
- [ ] OPNsense firewall logs
- [ ] NTP server (chrony) logs
- [ ] Gotra application server
  - [ ] Gunicorn access/error logs
  - [ ] Celery worker logs
  - [ ] Redis logs

### Dashboard Improvements
- [ ] NTP Time Sync Dashboard
- [ ] Firewall Security Dashboard
- [ ] Application Performance Dashboard
- [ ] Home page with health overview

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
- [ ] Node exporter on all Linux hosts
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
- [ ] Enable Loki retention policy (currently unlimited)
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
| Log ingestion rate | >1000 msg/min | TBD |
| Dashboard load time | <3 seconds | TBD |
| Alert response time | <5 minutes | N/A |
| Log retention | 30 days | Unlimited |
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
