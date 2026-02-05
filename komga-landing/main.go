package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SystemStats struct {
	Hostname    string
	Uptime      string
	CPUPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	MemPercent  float64
	DiskUsed    uint64
	DiskTotal   uint64
	DiskPercent float64
	LoadAvg     string
	OS          string
	Kernel      string
}

type KomgaStats struct {
	Libraries int
	Series    int
	Books     int
}

type HistoricalPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

type PageData struct {
	System     SystemStats
	Komga      KomgaStats
	CPUHistory []HistoricalPoint
	MemHistory []HistoricalPoint
	Updated    string
}

func getSystemStats() SystemStats {
	stats := SystemStats{}

	hostname, _ := os.Hostname()
	stats.Hostname = hostname

	uptime, _ := os.ReadFile("/proc/uptime")
	if len(uptime) > 0 {
		secs, _ := strconv.ParseFloat(strings.Split(string(uptime), " ")[0], 64)
		days := int(secs / 86400)
		hours := int(secs/3600) % 24
		mins := int(secs/60) % 60
		stats.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}

	cmd := exec.Command("sh", "-c", "grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage}'")
	out, _ := cmd.Output()
	stats.CPUPercent, _ = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)

	meminfo, _ := os.ReadFile("/proc/meminfo")
	lines := strings.Split(string(meminfo), "\n")
	var memTotal, memAvail uint64
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %d kB", &memTotal)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %d kB", &memAvail)
		}
	}
	stats.MemTotal = memTotal * 1024
	stats.MemUsed = (memTotal - memAvail) * 1024
	if memTotal > 0 {
		stats.MemPercent = float64(memTotal-memAvail) / float64(memTotal) * 100
	}

	cmd = exec.Command("df", "-B1", "/")
	out, _ = cmd.Output()
	lines = strings.Split(string(out), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			stats.DiskTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			stats.DiskUsed, _ = strconv.ParseUint(fields[2], 10, 64)
			if stats.DiskTotal > 0 {
				stats.DiskPercent = float64(stats.DiskUsed) / float64(stats.DiskTotal) * 100
			}
		}
	}

	loadavg, _ := os.ReadFile("/proc/loadavg")
	parts := strings.Split(string(loadavg), " ")
	if len(parts) >= 3 {
		stats.LoadAvg = strings.Join(parts[:3], " ")
	}

	osRelease, _ := os.ReadFile("/etc/os-release")
	for _, line := range strings.Split(string(osRelease), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			stats.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}

	cmd = exec.Command("uname", "-r")
	out, _ = cmd.Output()
	stats.Kernel = strings.TrimSpace(string(out))

	return stats
}

func getKomgaStats() KomgaStats {
	stats := KomgaStats{}
	cmd := exec.Command("docker", "exec", "komga", "ls", "-la", "/config")
	cmd.Output()
	return stats
}

func getHistoricalData(query string) []HistoricalPoint {
	points := []HistoricalPoint{}

	promURL := fmt.Sprintf("https://prometheus.sentinella.alpina/api/v1/query_range?query=%s&start=%d&end=%d&step=3600",
		query, time.Now().Add(-30*24*time.Hour).Unix(), time.Now().Unix())

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", promURL, nil)
	req.SetBasicAuth("admin", "vURLumGa0GMu4/nR2+vejcenAQBqt1un")

	resp, err := client.Do(req)
	if err != nil {
		return points
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if results, ok := data["result"].([]interface{}); ok && len(results) > 0 {
			if first, ok := results[0].(map[string]interface{}); ok {
				if values, ok := first["values"].([]interface{}); ok {
					for _, v := range values {
						if pair, ok := v.([]interface{}); ok && len(pair) == 2 {
							ts := int64(pair[0].(float64))
							val, _ := strconv.ParseFloat(pair[1].(string), 64)
							points = append(points, HistoricalPoint{
								Time:  time.Unix(ts, 0).Format("Jan 2"),
								Value: val,
							})
						}
					}
				}
			}
		}
	}

	if len(points) > 60 {
		step := len(points) / 60
		sampled := []HistoricalPoint{}
		for i := 0; i < len(points); i += step {
			sampled = append(sampled, points[i])
		}
		return sampled
	}

	return points
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func main() {
	tmpl := template.Must(template.New("index").Funcs(template.FuncMap{
		"formatBytes": formatBytes,
		"printf":      fmt.Sprintf,
	}).Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cpuQuery := `100-avg(rate(node_cpu_seconds_total{instance="komga.alpina:9100",mode="idle"}[5m]))*100`
		memQuery := `(1-node_memory_MemAvailable_bytes{instance="komga.alpina:9100"}/node_memory_MemTotal_bytes{instance="komga.alpina:9100"})*100`

		data := PageData{
			System:     getSystemStats(),
			Komga:      getKomgaStats(),
			CPUHistory: getHistoricalData(cpuQuery),
			MemHistory: getHistoricalData(memQuery),
			Updated:    time.Now().Format("2006-01-02 15:04:05"),
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"system": getSystemStats(),
			"komga":  getKomgaStats(),
		})
	})

	fmt.Println("Komga Landing Page running on :80")
	http.ListenAndServe(":80", nil)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Komga Server</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #e4e4e4;
            min-height: 100vh;
            padding: 2rem;
        }
        .container { max-width: 1400px; margin: 0 auto; }
        header {
            text-align: center;
            margin-bottom: 3rem;
            padding: 2rem;
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.1);
        }
        h1 {
            font-size: 3rem;
            background: linear-gradient(120deg, #e94560, #ff6b6b);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 0.5rem;
        }
        .subtitle { color: #888; font-size: 1.1rem; }
        .hostname { color: #e94560; font-weight: 600; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
        .card {
            background: rgba(255,255,255,0.05);
            border-radius: 16px;
            padding: 1.5rem;
            border: 1px solid rgba(255,255,255,0.1);
            backdrop-filter: blur(10px);
            transition: transform 0.3s, box-shadow 0.3s;
        }
        .card:hover {
            transform: translateY(-5px);
            box-shadow: 0 20px 40px rgba(0,0,0,0.3);
        }
        .card-title {
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #888;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .card-title::before { content: ""; display: inline-block; width: 8px; height: 8px; background: #e94560; border-radius: 50%; }
        .stat-value {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(120deg, #fff, #ccc);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .stat-label { color: #666; margin-top: 0.25rem; }
        .progress-bar {
            height: 8px;
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
            overflow: hidden;
            margin-top: 1rem;
        }
        .progress-fill {
            height: 100%;
            border-radius: 4px;
            transition: width 0.5s ease;
        }
        .progress-fill.cpu { background: linear-gradient(90deg, #00c9ff, #92fe9d); }
        .progress-fill.mem { background: linear-gradient(90deg, #f093fb, #f5576c); }
        .progress-fill.disk { background: linear-gradient(90deg, #4facfe, #00f2fe); }
        .chart-container {
            background: rgba(255,255,255,0.05);
            border-radius: 16px;
            padding: 1.5rem;
            border: 1px solid rgba(255,255,255,0.1);
            margin-bottom: 2rem;
        }
        .chart-title { font-size: 1.2rem; margin-bottom: 1rem; color: #fff; }
        .info-grid { display: grid; grid-template-columns: 1fr; gap: 0.5rem; }
        .info-item { display: flex; justify-content: space-between; padding: 0.5rem 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
        .info-label { color: #888; }
        .info-value { color: #fff; font-weight: 500; }
        .komga-section {
            background: linear-gradient(135deg, rgba(233,69,96,0.1), rgba(255,107,107,0.05));
            border: 1px solid rgba(233,69,96,0.3);
        }
        footer {
            text-align: center;
            padding: 2rem;
            color: #666;
            font-size: 0.9rem;
        }
        .btn {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background: linear-gradient(120deg, #e94560, #ff6b6b);
            color: #fff;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .btn:hover { transform: translateY(-2px); box-shadow: 0 10px 20px rgba(233,69,96,0.3); }
        .charts-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 1.5rem; }
        @media (max-width: 768px) {
            h1 { font-size: 2rem; }
            .stat-value { font-size: 1.8rem; }
            .charts-grid { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📚 Komga Server</h1>
            <p class="subtitle">Comic & Manga Library • <span class="hostname">{{.System.Hostname}}</span></p>
            <p style="margin-top: 1rem;">
                <a href="http://komga.alpina:25600" class="btn" target="_blank">Open Komga →</a>
            </p>
        </header>

        <div class="grid">
            <div class="card">
                <div class="card-title">CPU Usage</div>
                <div class="stat-value">{{printf "%.1f" .System.CPUPercent}}%</div>
                <div class="stat-label">Load: {{.System.LoadAvg}}</div>
                <div class="progress-bar"><div class="progress-fill cpu" style="width: {{printf "%.0f" .System.CPUPercent}}%"></div></div>
            </div>
            <div class="card">
                <div class="card-title">Memory</div>
                <div class="stat-value">{{printf "%.1f" .System.MemPercent}}%</div>
                <div class="stat-label">{{formatBytes .System.MemUsed}} / {{formatBytes .System.MemTotal}}</div>
                <div class="progress-bar"><div class="progress-fill mem" style="width: {{printf "%.0f" .System.MemPercent}}%"></div></div>
            </div>
            <div class="card">
                <div class="card-title">Disk Usage</div>
                <div class="stat-value">{{printf "%.1f" .System.DiskPercent}}%</div>
                <div class="stat-label">{{formatBytes .System.DiskUsed}} / {{formatBytes .System.DiskTotal}}</div>
                <div class="progress-bar"><div class="progress-fill disk" style="width: {{printf "%.0f" .System.DiskPercent}}%"></div></div>
            </div>
            <div class="card">
                <div class="card-title">System Info</div>
                <div class="info-grid">
                    <div class="info-item"><span class="info-label">Uptime</span><span class="info-value">{{.System.Uptime}}</span></div>
                    <div class="info-item"><span class="info-label">OS</span><span class="info-value">{{.System.OS}}</span></div>
                    <div class="info-item"><span class="info-label">Kernel</span><span class="info-value">{{.System.Kernel}}</span></div>
                </div>
            </div>
        </div>

        <div class="charts-grid">
            <div class="chart-container">
                <div class="chart-title">📊 CPU Usage (30 Days)</div>
                <canvas id="cpuChart" height="120"></canvas>
            </div>
            <div class="chart-container">
                <div class="chart-title">📊 Memory Usage (30 Days)</div>
                <canvas id="memChart" height="120"></canvas>
            </div>
        </div>

        <footer>
            <p>Last updated: {{.Updated}} • Powered by Go</p>
        </footer>
    </div>

    <script>
        const cpuData = {{.CPUHistory}};
        const memData = {{.MemHistory}};

        const chartOptions = {
            responsive: true,
            maintainAspectRatio: true,
            plugins: { legend: { display: false } },
            scales: {
                x: { grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#888', maxTicksLimit: 8 } },
                y: { grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#888' }, min: 0, max: 100 }
            }
        };

        if (cpuData && cpuData.length > 0) {
            new Chart(document.getElementById('cpuChart'), {
                type: 'line',
                data: {
                    labels: cpuData.map(p => p.time),
                    datasets: [{
                        data: cpuData.map(p => p.value),
                        borderColor: '#00c9ff',
                        backgroundColor: 'rgba(0,201,255,0.1)',
                        fill: true,
                        tension: 0.4,
                        pointRadius: 0
                    }]
                },
                options: chartOptions
            });
        }

        if (memData && memData.length > 0) {
            new Chart(document.getElementById('memChart'), {
                type: 'line',
                data: {
                    labels: memData.map(p => p.time),
                    datasets: [{
                        data: memData.map(p => p.value),
                        borderColor: '#f5576c',
                        backgroundColor: 'rgba(245,87,108,0.1)',
                        fill: true,
                        tension: 0.4,
                        pointRadius: 0
                    }]
                },
                options: chartOptions
            });
        }
    </script>
</body>
</html>`
