package main

import (
	"bufio"
	"crypto/tls"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed template.html
var htmlTemplate string

var _ embed.FS

// SystemStats holds system resource information
type SystemStats struct {
	Hostname    string  `json:"hostname"`
	Uptime      string  `json:"uptime"`
	UptimeSecs  float64 `json:"uptimeSecs"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemTotal    uint64  `json:"memTotal"`
	MemUsed     uint64  `json:"memUsed"`
	MemPercent  float64 `json:"memPercent"`
	DiskTotal   string  `json:"diskTotal"`
	DiskUsed    string  `json:"diskUsed"`
	DiskPercent string  `json:"diskPercent"`
	LoadAvg     string  `json:"loadAvg"`
	OS          string  `json:"os"`
	Kernel      string  `json:"kernel"`
}

// NTPSource represents a single NTP source from chronyc sources
type NTPSource struct {
	StatusIcon string `json:"statusIcon"`
	Name       string `json:"name"`
	Stratum    string `json:"stratum"`
	Poll       string `json:"poll"`
	Reach      string `json:"reach"`
	ReachBits  []string `json:"reachBits"`
	LastRx     string `json:"lastRx"`
	Offset     string `json:"offset"`
	FreqSkew   string `json:"freqSkew"`
	StdDev     string `json:"stdDev"`
	NTS        bool   `json:"nts"`
	Selected   bool   `json:"selected"`
}

// NTSDetail represents NTS authentication details for a source
type NTSDetail struct {
	Name         string `json:"name"`
	KeyLength    string `json:"keyLength"`
	LastAuth     string `json:"lastAuth"`
	Cookies      string `json:"cookies"`
	CookieLength string `json:"cookieLength"`
}

// NTPStats holds all NTP-related statistics
type NTPStats struct {
	Stratum       string      `json:"stratum"`
	RefID         string      `json:"refID"`
	Offset        float64     `json:"offset"`
	OffsetDisplay string      `json:"offsetDisplay"`
	FreqPPM       float64     `json:"freqPPM"`
	FreqDisplay   string      `json:"freqDisplay"`
	RootDelay     string      `json:"rootDelay"`
	RootDisp      string      `json:"rootDisp"`
	UpdateInt     string      `json:"updateInt"`
	SystemTime    string      `json:"systemTime"`
	LastOffset    string      `json:"lastOffset"`
	RMSOffset     string      `json:"rmsOffset"`
	ResidualFreq  string      `json:"residualFreq"`
	Skew          string      `json:"skew"`
	NTSCount      int         `json:"ntsCount"`
	TotalSources  int         `json:"totalSources"`
	OnlineSources int         `json:"onlineSources"`
	Sources       []NTPSource `json:"sources"`
	NTSDetails    []NTSDetail `json:"ntsDetails"`
	Synced        bool        `json:"synced"`
}

// ChartPoint is a single data point for charts
type ChartPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// ChartDataSet holds chart data for all metric types
type ChartDataSet struct {
	Offset []ChartPoint `json:"offset"`
	Freq   []ChartPoint `json:"freq"`
	MaxErr []ChartPoint `json:"maxErr"`
	EstErr []ChartPoint `json:"estErr"`
	PLL    []ChartPoint `json:"pll"`
}

// PageData is the top-level struct passed to the template
type PageData struct {
	NTP        NTPStats       `json:"ntp"`
	System     SystemStats    `json:"system"`
	Charts     ChartDataSet   `json:"charts"`
	ChartsJSON template.JS    `json:"-"`
	CPUJSON    template.JS    `json:"-"`
	MemJSON    template.JS    `json:"-"`
	UpdatedAt  string         `json:"updatedAt"`
}

func getSystemStats() SystemStats {
	var stats SystemStats

	hostname, err := os.Hostname()
	if err == nil {
		stats.Hostname = hostname
	}

	// Read uptime
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err == nil {
		fields := strings.Fields(string(uptimeData))
		if len(fields) >= 1 {
			secs, _ := strconv.ParseFloat(fields[0], 64)
			stats.UptimeSecs = secs
			days := int(secs) / 86400
			hours := (int(secs) % 86400) / 3600
			mins := (int(secs) % 3600) / 60
			if days > 0 {
				stats.Uptime = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
			} else if hours > 0 {
				stats.Uptime = fmt.Sprintf("%dh %dm", hours, mins)
			} else {
				stats.Uptime = fmt.Sprintf("%dm", mins)
			}
		}
	}

	// Read CPU stats
	statData, err := os.ReadFile("/proc/stat")
	if err == nil {
		lines := strings.Split(string(statData), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
			fields := strings.Fields(lines[0])
			if len(fields) >= 5 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				iowait := 0.0
				if len(fields) >= 6 {
					iowait, _ = strconv.ParseFloat(fields[5], 64)
				}
				total := user + nice + system + idle + iowait
				if total > 0 {
					stats.CPUPercent = math.Round(((total-idle)/total)*1000) / 10
				}
			}
		}
	}

	// Read memory info
	memData, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		var memTotal, memAvail uint64
		scanner := bufio.NewScanner(strings.NewReader(string(memData)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d kB", &memTotal)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d kB", &memAvail)
			}
		}
		stats.MemTotal = memTotal * 1024
		stats.MemUsed = (memTotal - memAvail) * 1024
		if memTotal > 0 {
			stats.MemPercent = math.Round(float64(memTotal-memAvail) / float64(memTotal) * 1000) / 10
		}
	}

	// Disk usage
	dfOut, err := exec.Command("df", "-h", "/").Output()
	if err == nil {
		lines := strings.Split(string(dfOut), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				stats.DiskTotal = fields[1]
				stats.DiskUsed = fields[2]
				stats.DiskPercent = fields[4]
			}
		}
	}

	// Load average
	loadData, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		fields := strings.Fields(string(loadData))
		if len(fields) >= 3 {
			stats.LoadAvg = strings.Join(fields[:3], " ")
		}
	}

	// OS info
	osRelease, err := os.ReadFile("/etc/os-release")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(osRelease)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				stats.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}

	// Kernel
	kernelOut, err := exec.Command("uname", "-r").Output()
	if err == nil {
		stats.Kernel = strings.TrimSpace(string(kernelOut))
	}

	return stats
}

func reachToBits(reach string) []string {
	val, err := strconv.ParseUint(strings.TrimSpace(reach), 8, 16)
	if err != nil {
		return []string{"0", "0", "0", "0", "0", "0", "0", "0"}
	}
	bits := fmt.Sprintf("%08b", val)
	result := make([]string, 8)
	for i, c := range bits {
		result[i] = string(c)
	}
	return result
}

func getNTSMap() map[string]bool {
	ntsMap := make(map[string]bool)
	out, err := exec.Command("chronyc", "authdata").Output()
	if err != nil {
		return ntsMap
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "NTS" {
			ntsMap[fields[0]] = true
		}
	}
	return ntsMap
}

func getNTSDetails() []NTSDetail {
	var details []NTSDetail
	out, err := exec.Command("chronyc", "authdata").Output()
	if err != nil {
		return details
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 6 && fields[1] == "NTS" {
			detail := NTSDetail{
				Name:         fields[0],
				KeyLength:    fields[4],
				LastAuth:     fields[5],
				Cookies:      "",
				CookieLength: "",
			}
			if len(fields) >= 9 {
				detail.Cookies = fields[8]
			}
			if len(fields) >= 10 {
				detail.CookieLength = fields[9]
			}
			details = append(details, detail)
		}
	}
	return details
}

func getSourceStats() map[string][2]string {
	result := make(map[string][2]string)
	out, err := exec.Command("chronyc", "sourcestats").Output()
	if err != nil {
		return result
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "=") || strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "210") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 8 {
			name := fields[0]
			freqSkew := fields[5]
			stdDev := fields[7]
			result[name] = [2]string{freqSkew, stdDev}
		}
	}
	return result
}

func getNTPStats() NTPStats {
	var stats NTPStats
	ntsMap := getNTSMap()
	ntsDetails := getNTSDetails()
	sourceStatsMap := getSourceStats()

	stats.NTSDetails = ntsDetails

	// Parse chronyc sources
	out, err := exec.Command("chronyc", "sources").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "=") || strings.HasPrefix(line, "MS") || strings.HasPrefix(line, "210") {
				continue
			}
			if len(line) < 3 {
				continue
			}
			modeChar := string(line[0])
			stateChar := string(line[1])
			_ = modeChar

			rest := strings.TrimSpace(line[2:])
			fields := strings.Fields(rest)
			if len(fields) < 7 {
				continue
			}

			name := fields[0]
			icon := stateChar
			switch stateChar {
			case "*":
				icon = "*"
			case "+":
				icon = "+"
			case "-":
				icon = "-"
			case "x":
				icon = "x"
			case "?":
				icon = "?"
			case "~":
				icon = "~"
			}

			selected := stateChar == "*"

			source := NTPSource{
				StatusIcon: icon,
				Name:       name,
				Stratum:    fields[1],
				Poll:       fields[2],
				Reach:      fields[3],
				ReachBits:  reachToBits(fields[3]),
				LastRx:     fields[4],
				Offset:     fields[5],
				NTS:        ntsMap[name],
				Selected:   selected,
			}

			if ss, ok := sourceStatsMap[name]; ok {
				source.FreqSkew = ss[0]
				source.StdDev = ss[1]
			}

			stats.Sources = append(stats.Sources, source)
			stats.TotalSources++
			if ntsMap[name] {
				stats.NTSCount++
			}
		}
	}

	// Parse chronyc activity
	actOut, err := exec.Command("chronyc", "activity").Output()
	if err == nil {
		lines := strings.Split(string(actOut), "\n")
		for _, line := range lines {
			if strings.Contains(line, "online") {
				fields := strings.Fields(line)
				if len(fields) >= 1 {
					n, _ := strconv.Atoi(fields[0])
					stats.OnlineSources = n
				}
				break
			}
		}
	}

	// Parse chronyc tracking
	trackOut, err := exec.Command("chronyc", "tracking").Output()
	if err == nil {
		lines := strings.Split(string(trackOut), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "Reference ID":
				stats.RefID = val
			case "Stratum":
				stats.Stratum = val
			case "System time":
				stats.SystemTime = val
				re := regexp.MustCompile(`([0-9.]+)\s+seconds\s+(slow|fast)`)
				m := re.FindStringSubmatch(val)
				if len(m) >= 3 {
					offset, _ := strconv.ParseFloat(m[1], 64)
					if m[2] == "slow" {
						offset = -offset
					}
					stats.Offset = offset
				}
			case "Last offset":
				stats.LastOffset = val
			case "RMS offset":
				stats.RMSOffset = val
			case "Frequency":
				stats.FreqDisplay = val
				re := regexp.MustCompile(`([0-9.]+)\s+ppm`)
				m := re.FindStringSubmatch(val)
				if len(m) >= 2 {
					ppm, _ := strconv.ParseFloat(m[1], 64)
					if strings.Contains(val, "slow") {
						ppm = -ppm
					}
					stats.FreqPPM = ppm
				}
			case "Residual freq":
				stats.ResidualFreq = val
			case "Skew":
				stats.Skew = val
			case "Root delay":
				stats.RootDelay = val
			case "Root dispersion":
				stats.RootDisp = val
			case "Update interval":
				stats.UpdateInt = val
			case "Leap status":
				stats.Synced = val == "Normal"
			}
		}
	}

	// Format offset display
	absOffset := math.Abs(stats.Offset)
	if absOffset < 1e-6 {
		stats.OffsetDisplay = fmt.Sprintf("%.1f ns", stats.Offset*1e9)
	} else if absOffset < 1e-3 {
		stats.OffsetDisplay = fmt.Sprintf("%.1f us", stats.Offset*1e6)
	} else {
		stats.OffsetDisplay = fmt.Sprintf("%.3f ms", stats.Offset*1e3)
	}

	// Format chrony tracking values for display
	stats.RootDelay = formatSecondsStr(stats.RootDelay)
	stats.RootDisp = formatSecondsStr(stats.RootDisp)
	stats.UpdateInt = formatSecondsStr(stats.UpdateInt)
	stats.RMSOffset = formatSecondsStr(stats.RMSOffset)
	stats.LastOffset = formatSecondsStr(stats.LastOffset)
	stats.SystemTime = formatSecondsStr(stats.SystemTime)
	stats.ResidualFreq = formatFreqStr(stats.ResidualFreq)
	stats.Skew = formatFreqStr(stats.Skew)

	return stats
}

// formatSecondsStr parses "0.012600483 seconds" and returns "12.6 ms"
func formatSecondsStr(val string) string {
	re := regexp.MustCompile(`([+-]?[\d.]+)\s+seconds`)
	m := re.FindStringSubmatch(val)
	if len(m) < 2 {
		return val
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return val
	}
	abs := math.Abs(f)
	sign := ""
	if f < 0 {
		sign = "-"
	}
	rest := ""
	// Check for "slow"/"fast" suffix
	if strings.Contains(val, "slow") {
		rest = " slow"
	} else if strings.Contains(val, "fast") {
		rest = " fast"
	}
	if abs >= 60 {
		return fmt.Sprintf("%s%.1f min%s", sign, abs/60, rest)
	} else if abs >= 1 {
		return fmt.Sprintf("%s%.1f s%s", sign, abs, rest)
	} else if abs >= 0.001 {
		return fmt.Sprintf("%s%.2f ms%s", sign, abs*1000, rest)
	} else if abs >= 0.000001 {
		return fmt.Sprintf("%s%.1f us%s", sign, abs*1e6, rest)
	}
	return fmt.Sprintf("%s%.1f ns%s", sign, abs*1e9, rest)
}

// formatFreqStr formats "0.066 ppm" or "2.060 ppm" to fewer decimals
func formatFreqStr(val string) string {
	re := regexp.MustCompile(`([+-]?[\d.]+)\s+ppm`)
	m := re.FindStringSubmatch(val)
	if len(m) < 2 {
		return val
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return val
	}
	return fmt.Sprintf("%.2f ppm", f)
}

func fetchPromRangeWithFormat(query, start, end, step, timeFmt string) []ChartPoint {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	promURL := fmt.Sprintf("https://prometheus.sentinella.alpina/api/v1/query_range?query=%s&start=%s&end=%s&step=%s",
		url.QueryEscape(query), start, end, step)

	req, err := http.NewRequest("GET", promURL, nil)
	if err != nil {
		return nil
	}
	req.SetBasicAuth("admin", "vURLumGa0GMu4/nR2+vejcenAQBqt1un")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Values [][]interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil
	}

	if promResp.Status != "success" || len(promResp.Data.Result) == 0 {
		return nil
	}

	values := promResp.Data.Result[0].Values
	maxPoints := 200
	skipEvery := 1
	if len(values) > maxPoints {
		skipEvery = len(values) / maxPoints
	}

	var points []ChartPoint
	for i, v := range values {
		if skipEvery > 1 && i%skipEvery != 0 && i != len(values)-1 {
			continue
		}
		if len(v) < 2 {
			continue
		}
		ts, ok := v[0].(float64)
		if !ok {
			continue
		}
		valStr, ok := v[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		t := time.Unix(int64(ts), 0)
		points = append(points, ChartPoint{
			Time:  t.Format(timeFmt),
			Value: val,
		})
	}
	return points
}

func fetchChartSetFull(rangeName string) ChartDataSet {
	var ds ChartDataSet
	now := time.Now()

	var duration time.Duration
	var step string
	var timeFmt string

	switch rangeName {
	case "1h":
		duration = 1 * time.Hour
		step = "30"
		timeFmt = "15:04:05"
	case "6h":
		duration = 6 * time.Hour
		step = "120"
		timeFmt = "15:04"
	case "24h":
		duration = 24 * time.Hour
		step = "300"
		timeFmt = "15:04"
	case "7d":
		duration = 7 * 24 * time.Hour
		step = "1800"
		timeFmt = "Mon 15h"
	case "30d":
		duration = 30 * 24 * time.Hour
		step = "7200"
		timeFmt = "Jan 2"
	default:
		duration = 24 * time.Hour
		step = "300"
		timeFmt = "15:04"
	}

	startStr := fmt.Sprintf("%d", now.Add(-duration).Unix())
	endStr := fmt.Sprintf("%d", now.Unix())

	type result struct {
		name   string
		points []ChartPoint
	}

	var wg sync.WaitGroup
	ch := make(chan result, 5)

	inst := "ntp.alpina:9100"
	queries := map[string]string{
		"offset": fmt.Sprintf("node_timex_offset_seconds{instance=\"%s\"} * 1e6", inst),
		"freq":   fmt.Sprintf("(node_timex_frequency_adjustment_ratio{instance=\"%s\"} - 1) * 1e6", inst),
		"maxErr": fmt.Sprintf("node_timex_maxerror_seconds{instance=\"%s\"} * 1e6", inst),
		"estErr": fmt.Sprintf("node_timex_estimated_error_seconds{instance=\"%s\"} * 1e6", inst),
		"pll":    fmt.Sprintf("node_timex_loop_time_constant{instance=\"%s\"}", inst),
	}

	for name, query := range queries {
		wg.Add(1)
		go func(n, q string) {
			defer wg.Done()
			points := fetchPromRangeWithFormat(q, startStr, endStr, step, timeFmt)
			ch <- result{name: n, points: points}
		}(name, query)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		switch r.name {
		case "offset":
			ds.Offset = r.points
		case "freq":
			ds.Freq = r.points
		case "maxErr":
			ds.MaxErr = r.points
		case "estErr":
			ds.EstErr = r.points
		case "pll":
			ds.PLL = r.points
		}
	}

	return ds
}

func fetchCPU30d() []ChartPoint {
	now := time.Now()
	start := fmt.Sprintf("%d", now.Add(-30*24*time.Hour).Unix())
	end := fmt.Sprintf("%d", now.Unix())
	return fetchPromRangeWithFormat(
		"100-(avg(rate(node_cpu_seconds_total{instance=\"ntp.alpina:9100\",mode=\"idle\"}[5m]))*100)",
		start, end, "7200", "Jan 2",
	)
}

func fetchMem30d() []ChartPoint {
	now := time.Now()
	start := fmt.Sprintf("%d", now.Add(-30*24*time.Hour).Unix())
	end := fmt.Sprintf("%d", now.Unix())
	return fetchPromRangeWithFormat(
		"(1-node_memory_MemAvailable_bytes{instance=\"ntp.alpina:9100\"}/node_memory_MemTotal_bytes{instance=\"ntp.alpina:9100\"})*100",
		start, end, "7200", "Jan 2",
	)
}

func main() {
	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		ntpStats := getNTPStats()
		sysStats := getSystemStats()
		charts := fetchChartSetFull("24h")
		cpuData := fetchCPU30d()
		memData := fetchMem30d()

		chartsJSON, _ := json.Marshal(charts)
		cpuJSON, _ := json.Marshal(cpuData)
		memJSON, _ := json.Marshal(memData)

		data := PageData{
			NTP:        ntpStats,
			System:     sysStats,
			Charts:     charts,
			ChartsJSON: template.JS(chartsJSON),
			CPUJSON:    template.JS(cpuJSON),
			MemJSON:    template.JS(memJSON),
			UpdatedAt:  time.Now().Format("2006-01-02 15:04:05 MST"),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, "Internal Server Error", 500)
		}
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		ntpStats := getNTPStats()
		sysStats := getSystemStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ntp":    ntpStats,
			"system": sysStats,
		})
	})

	http.HandleFunc("/api/charts", func(w http.ResponseWriter, r *http.Request) {
		rangeName := r.URL.Query().Get("range")
		if rangeName == "" {
			rangeName = "24h"
		}
		charts := fetchChartSetFull(rangeName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(charts)
	})

	log.Println("NTP Landing Page listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
