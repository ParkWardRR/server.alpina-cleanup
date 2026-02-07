package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

const baseURL = "http://ntp.alpina"

func TestLandingPage(t *testing.T) {
	// Verify server is reachable
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		t.Fatalf("Cannot reach %s: %v", baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	// Launch playwright
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("Could not start Playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("Could not create page: %v", err)
	}

	_, err = page.Goto(baseURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		t.Fatalf("Could not navigate to %s: %v", baseURL, err)
	}

	// --- Test 1: Header loads ---
	t.Run("Header_renders", func(t *testing.T) {
		title, err := page.Locator("h1").TextContent()
		if err != nil {
			t.Fatalf("Could not get h1 text: %v", err)
		}
		if !strings.Contains(title, "NTP Server") {
			t.Errorf("Expected h1 to contain 'NTP Server', got: %s", title)
		}
	})

	// --- Test 2: Sync badge shows Synchronized ---
	t.Run("Sync_badge", func(t *testing.T) {
		badge, err := page.Locator(".badge-green").TextContent()
		if err != nil {
			t.Fatalf("Could not find green badge: %v", err)
		}
		if !strings.Contains(badge, "Synchronized") {
			t.Errorf("Expected Synchronized badge, got: %s", badge)
		}
	})

	// --- Test 3: NTS badge shows 7 ---
	t.Run("NTS_badge", func(t *testing.T) {
		badge, err := page.Locator(".badge-blue").TextContent()
		if err != nil {
			t.Fatalf("Could not find blue badge: %v", err)
		}
		if !strings.Contains(badge, "7") || !strings.Contains(badge, "Authenticated") {
			t.Errorf("Expected NTS badge with 7 authenticated, got: %s", badge)
		}
	})

	// --- Test 4: Hero metric values have NO raw "seconds" ---
	t.Run("No_raw_seconds_in_hero", func(t *testing.T) {
		cards, err := page.Locator(".hero-card .value").All()
		if err != nil {
			t.Fatalf("Could not get hero cards: %v", err)
		}
		for _, card := range cards {
			text, _ := card.TextContent()
			if strings.Contains(text, "seconds") {
				t.Errorf("Hero card still has raw 'seconds': %s", text)
			}
			// Check for excessive decimals (more than 3 decimal places)
			parts := strings.Split(text, ".")
			if len(parts) == 2 {
				// Strip non-numeric suffix
				decimal := parts[1]
				numDigits := 0
				for _, c := range decimal {
					if c >= '0' && c <= '9' {
						numDigits++
					} else {
						break
					}
				}
				if numDigits > 3 {
					t.Errorf("Hero card has too many decimals (%d): %s", numDigits, text)
				}
			}
		}
	})

	// --- Test 5: Charts have data (canvas elements have rendered) ---
	t.Run("Charts_have_data", func(t *testing.T) {
		// Wait a moment for Chart.js to render
		page.WaitForTimeout(2000)

		chartIDs := []string{"chartOffset", "chartFreq", "chartError", "chartPLL"}
		for _, id := range chartIDs {
			// Check the chart canvas exists
			canvas := page.Locator("#" + id)
			visible, err := canvas.IsVisible()
			if err != nil {
				t.Errorf("Could not check visibility of #%s: %v", id, err)
				continue
			}
			if !visible {
				t.Errorf("Chart canvas #%s is not visible", id)
				continue
			}

			// Check that Chart.js actually rendered data by checking the instance
			hasData, err := page.Evaluate(fmt.Sprintf(
				"typeof chartInstances !== 'undefined' && chartInstances['%s'] && chartInstances['%s'].data.datasets[0].data.length > 0",
				id, id))
			if err != nil {
				t.Errorf("Could not evaluate chart data for #%s: %v", id, err)
				continue
			}
			if hasData != true {
				t.Errorf("Chart #%s has no data rendered", id)
			}
		}
	})

	// --- Test 6: Sources table has rows ---
	t.Run("Sources_table_populated", func(t *testing.T) {
		rows, err := page.Locator("table tbody tr").All()
		if err != nil {
			t.Fatalf("Could not get table rows: %v", err)
		}
		if len(rows) < 10 {
			t.Errorf("Expected at least 10 source rows, got %d", len(rows))
		}
	})

	// --- Test 7: NTS badges appear in table ---
	t.Run("NTS_badges_in_table", func(t *testing.T) {
		badges, err := page.Locator("table .nts-badge").All()
		if err != nil {
			t.Fatalf("Could not find NTS badges: %v", err)
		}
		if len(badges) < 7 {
			t.Errorf("Expected at least 7 NTS badges in table, got %d", len(badges))
		}
	})

	// --- Test 8: Reach dots render ---
	t.Run("Reach_dots_render", func(t *testing.T) {
		dots, err := page.Locator(".reach-dot").All()
		if err != nil {
			t.Fatalf("Could not find reach dots: %v", err)
		}
		// 33 sources * 8 dots each = 264
		if len(dots) < 100 {
			t.Errorf("Expected many reach dots, got %d", len(dots))
		}
	})

	// --- Test 9: Performance Profile section exists ---
	t.Run("Performance_profile", func(t *testing.T) {
		text, err := page.Locator(".profile-card").TextContent()
		if err != nil {
			t.Fatalf("Could not find profile card: %v", err)
		}
		if !strings.Contains(text, "SCHED_FIFO") {
			t.Errorf("Performance profile missing SCHED_FIFO")
		}
		if !strings.Contains(text, "Kernel Tuning") {
			t.Errorf("Performance profile missing Kernel Tuning")
		}
	})

	// --- Test 10: NTS Authentication table ---
	t.Run("NTS_auth_table", func(t *testing.T) {
		ntsRows, err := page.Locator(".nts-table tbody tr").All()
		if err != nil {
			t.Fatalf("Could not find NTS auth table rows: %v", err)
		}
		if len(ntsRows) < 7 {
			t.Errorf("Expected 7 NTS auth rows, got %d", len(ntsRows))
		}
	})

	// --- Test 11: Chrony tracking info has formatted values ---
	t.Run("Tracking_formatted_values", func(t *testing.T) {
		infoRows, err := page.Locator(".info-row .info-val").All()
		if err != nil {
			t.Fatalf("Could not get info-row values: %v", err)
		}
		for _, row := range infoRows {
			text, _ := row.TextContent()
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			// No raw "seconds" with many decimal places
			if strings.Contains(text, "seconds") && strings.Count(text, ".") > 0 {
				parts := strings.Split(text, ".")
				if len(parts) >= 2 {
					numDigits := 0
					for _, c := range parts[1] {
						if c >= '0' && c <= '9' {
							numDigits++
						} else {
							break
						}
					}
					if numDigits > 4 {
						t.Errorf("Info row has too many decimals: %s", text)
					}
				}
			}
		}
	})

	// --- Test 12: Tab switching works ---
	t.Run("Chart_tab_switching", func(t *testing.T) {
		tab1h := page.Locator(".chart-tab[data-range='1h']")
		err := tab1h.Click()
		if err != nil {
			t.Fatalf("Could not click 1h tab: %v", err)
		}

		// Wait for fetch to complete
		page.WaitForTimeout(3000)

		// Check that the tab is now active
		cls, _ := tab1h.GetAttribute("class")
		if !strings.Contains(cls, "active") {
			t.Errorf("1h tab not active after click")
		}

		// Verify chart data updated
		hasData, err := page.Evaluate(
			"typeof chartInstances !== 'undefined' && chartInstances['chartOffset'] && chartInstances['chartOffset'].data.datasets[0].data.length > 0")
		if err != nil {
			t.Errorf("Could not check chart after tab switch: %v", err)
		}
		if hasData != true {
			t.Errorf("Chart data empty after switching to 1h tab")
		}
	})

	// --- Test 13: System resources section ---
	t.Run("System_resources", func(t *testing.T) {
		cpuText, err := page.Locator("#chartCPU").IsVisible()
		if err != nil || !cpuText {
			t.Errorf("CPU chart not visible")
		}
		memText, err := page.Locator("#chartMem").IsVisible()
		if err != nil || !memText {
			t.Errorf("Memory chart not visible")
		}
	})

	// --- Test 14: API endpoint ---
	t.Run("API_stats", func(t *testing.T) {
		apiResp, err := client.Get(baseURL + "/api/stats")
		if err != nil {
			t.Fatalf("API request failed: %v", err)
		}
		defer apiResp.Body.Close()

		var data map[string]interface{}
		json.NewDecoder(apiResp.Body).Decode(&data)

		ntp, ok := data["ntp"].(map[string]interface{})
		if !ok {
			t.Fatalf("API response missing ntp field")
		}

		total := int(ntp["totalSources"].(float64))
		if total < 30 {
			t.Errorf("Expected 30+ sources, got %d", total)
		}

		ntsCount := int(ntp["ntsCount"].(float64))
		if ntsCount != 7 {
			t.Errorf("Expected 7 NTS sources, got %d", ntsCount)
		}

		rootDelay := ntp["rootDelay"].(string)
		if strings.Contains(rootDelay, "seconds") {
			t.Errorf("API rootDelay still raw: %s", rootDelay)
		}
	})

	// --- Test 15: API charts endpoint ---
	t.Run("API_charts", func(t *testing.T) {
		apiResp, err := client.Get(baseURL + "/api/charts?range=1h")
		if err != nil {
			t.Fatalf("API charts request failed: %v", err)
		}
		defer apiResp.Body.Close()

		var data ChartDataSet
		json.NewDecoder(apiResp.Body).Decode(&data)

		if len(data.Offset) == 0 {
			t.Errorf("API charts returned no offset data")
		}
		if len(data.Freq) == 0 {
			t.Errorf("API charts returned no freq data")
		}
		if len(data.MaxErr) == 0 {
			t.Errorf("API charts returned no maxErr data")
		}
	})

	// Take a screenshot for verification
	page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("/tmp/ntp-landing-test.png"),
		FullPage: playwright.Bool(true),
	})
	t.Log("Screenshot saved to /tmp/ntp-landing-test.png")
}
