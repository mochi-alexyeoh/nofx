package news

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var defaultFeeds = []string{
	"https://cointelegraph.com/rss",
	"https://www.coindesk.com/arc/outboundfeeds/rss/",
	"https://www.theblock.co/rss.xml",
}

type Item struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Symbols     []string  `json:"symbols"`
	Sentiment   float64   `json:"sentiment"` // [-1,1]
}

type Config struct {
	EnableCryptoPanic       bool
	CryptoPanicAPIKey       string
	EnableAlphaVantage      bool
	AlphaVantageAPIKey      string
	UseAlphaVantageSentiment bool
	TickerScope             string // "auto" | "manual"
	ManualTickers           string // comma-separated
}

type Client struct {
	httpClient               *http.Client
	feeds                    []string
	enableCryptoPanic        bool
	cryptoPanicAPIKey        string
	enableAlphaVantage       bool
	alphaVantageAPIKey       string
	useAlphaVantageSentiment bool
	tickerScope              string
	manualTickers            string
}

func NewClient() *Client {
	return NewClientWithConfig(Config{})
}

func NewClientWithConfig(cfg Config) *Client {
	return &Client{
		httpClient:               &http.Client{Timeout: 8 * time.Second},
		feeds:                    defaultFeeds,
		enableCryptoPanic:        cfg.EnableCryptoPanic,
		cryptoPanicAPIKey:        strings.TrimSpace(cfg.CryptoPanicAPIKey),
		enableAlphaVantage:       cfg.EnableAlphaVantage,
		alphaVantageAPIKey:       strings.TrimSpace(cfg.AlphaVantageAPIKey),
		useAlphaVantageSentiment: cfg.UseAlphaVantageSentiment,
		tickerScope:              strings.ToLower(strings.TrimSpace(cfg.TickerScope)),
		manualTickers:            strings.TrimSpace(cfg.ManualTickers),
	}
}

func (c *Client) Fetch(symbols []string, lookbackHours int, maxItems int) ([]Item, error) {
	if lookbackHours <= 0 {
		lookbackHours = 12
	}
	if maxItems <= 0 {
		maxItems = 20
	}

	since := time.Now().Add(-time.Duration(lookbackHours) * time.Hour)
	targets := c.resolveTargets(symbols)

	items := make([]Item, 0, maxItems)
	seen := make(map[string]bool)

	for _, feed := range c.feeds {
		feedItems, err := c.fetchFeed(feed, targets, since)
		if err != nil {
			continue
		}
		for _, it := range feedItems {
			key := strings.ToLower(strings.TrimSpace(it.Title))
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, it)
		}
	}

	if c.enableCryptoPanic && c.cryptoPanicAPIKey != "" {
		cpItems, err := c.fetchCryptoPanic(targets, since)
		if err == nil {
			for _, it := range cpItems {
				key := strings.ToLower(strings.TrimSpace(it.Title))
				if key == "" || seen[key] {
					continue
				}
				seen[key] = true
				items = append(items, it)
			}
		}
	}

	if c.enableAlphaVantage && c.alphaVantageAPIKey != "" {
		avItems, err := c.fetchAlphaVantage(targets, since)
		if err == nil {
			for _, it := range avItems {
				key := strings.ToLower(strings.TrimSpace(it.Title))
				if key == "" || seen[key] {
					continue
				}
				seen[key] = true
				items = append(items, it)
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].PublishedAt.After(items[j].PublishedAt)
	})
	if len(items) > maxItems {
		items = items[:maxItems]
	}
	return items, nil
}

func (c *Client) fetchFeed(url string, targets map[string]bool, since time.Time) ([]Item, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "NOFX-NewsFetcher/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("feed status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}

	if strings.Contains(strings.ToLower(string(body[:min(200, len(body))])), "<feed") {
		return parseAtom(body, targets, since, url)
	}
	return parseRSS(body, targets, since, url)
}

func parseRSS(body []byte, targets map[string]bool, since time.Time, source string) ([]Item, error) {
	type rssItem struct {
		Title   string `xml:"title"`
		Link    string `xml:"link"`
		PubDate string `xml:"pubDate"`
	}
	type channel struct { Items []rssItem `xml:"item"` }
	type rss struct { Channel channel `xml:"channel"` }

	var r rss
	if err := xml.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	result := make([]Item, 0, len(r.Channel.Items))
	for _, it := range r.Channel.Items {
		t, ok := parseTime(it.PubDate)
		if !ok || t.Before(since) {
			continue
		}
		syms := matchSymbols(it.Title, targets)
		if len(syms) == 0 && !isRelevantHeadline(it.Title) {
			continue
		}
		result = append(result, Item{Title: strings.TrimSpace(it.Title), Link: strings.TrimSpace(it.Link), Source: sourceHost(source), PublishedAt: t, Symbols: syms, Sentiment: sentimentScore(it.Title)})
	}
	return result, nil
}

func parseAtom(body []byte, targets map[string]bool, since time.Time, source string) ([]Item, error) {
	type link struct { Href string `xml:"href,attr"` }
	type entry struct {
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Link    link   `xml:"link"`
	}
	type feed struct { Entries []entry `xml:"entry"` }

	var f feed
	if err := xml.Unmarshal(body, &f); err != nil {
		return nil, err
	}
	result := make([]Item, 0, len(f.Entries))
	for _, it := range f.Entries {
		t, ok := parseTime(it.Updated)
		if !ok || t.Before(since) {
			continue
		}
		syms := matchSymbols(it.Title, targets)
		if len(syms) == 0 && !isRelevantHeadline(it.Title) {
			continue
		}
		result = append(result, Item{Title: strings.TrimSpace(it.Title), Link: strings.TrimSpace(it.Link.Href), Source: sourceHost(source), PublishedAt: t, Symbols: syms, Sentiment: sentimentScore(it.Title)})
	}
	return result, nil
}

func (c *Client) fetchCryptoPanic(targets map[string]bool, since time.Time) ([]Item, error) {
	endpoint, _ := url.Parse("https://cryptopanic.com/api/v1/posts/")
	q := endpoint.Query()
	q.Set("auth_token", c.cryptoPanicAPIKey)
	q.Set("kind", "news")
	q.Set("public", "true")
	if len(targets) > 0 {
		coins := make([]string, 0, len(targets))
		for token := range targets {
			if len(token) <= 6 && token != "BITCOIN" && token != "ETHEREUM" && token != "RIPPLE" && token != "SOLANA" {
				coins = append(coins, token)
			}
		}
		sort.Strings(coins)
		if len(coins) > 0 {
			q.Set("currencies", strings.Join(coins, ","))
		}
	}
	endpoint.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	req.Header.Set("User-Agent", "NOFX-NewsFetcher/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cryptopanic status %d", resp.StatusCode)
	}

	var payload struct {
		Results []struct {
			Title       string `json:"title"`
			PublishedAt string `json:"published_at"`
			URL         string `json:"url"`
			Currencies  []struct {
				Code string `json:"code"`
			} `json:"currencies"`
			Source struct {
				Title string `json:"title"`
			} `json:"source"`
		} `json:"results"`
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	result := make([]Item, 0, len(payload.Results))
	for _, r := range payload.Results {
		t, ok := parseTime(r.PublishedAt)
		if !ok || t.Before(since) {
			continue
		}
		syms := make([]string, 0, len(r.Currencies))
		for _, ccy := range r.Currencies {
			cc := strings.ToUpper(strings.TrimSpace(ccy.Code))
			if cc != "" {
				syms = append(syms, cc)
			}
		}
		if len(syms) == 0 {
			syms = matchSymbols(r.Title, targets)
		}
		if len(syms) == 0 && !isRelevantHeadline(r.Title) {
			continue
		}
		source := strings.TrimSpace(r.Source.Title)
		if source == "" {
			source = "CryptoPanic"
		}
		result = append(result, Item{
			Title:       strings.TrimSpace(r.Title),
			Link:        strings.TrimSpace(r.URL),
			Source:      source,
			PublishedAt: t,
			Symbols:     syms,
			Sentiment:   sentimentScore(r.Title),
		})
	}
	return result, nil
}

func (c *Client) resolveTargets(symbols []string) map[string]bool {
	if c.tickerScope == "manual" && strings.TrimSpace(c.manualTickers) != "" {
		parts := strings.Split(c.manualTickers, ",")
		manual := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				manual = append(manual, p)
			}
		}
		if len(manual) > 0 {
			return normalizeTargets(manual)
		}
	}
	return normalizeTargets(symbols)
}

func (c *Client) fetchAlphaVantage(targets map[string]bool, since time.Time) ([]Item, error) {
	endpoint, _ := url.Parse("https://www.alphavantage.co/query")
	q := endpoint.Query()
	q.Set("function", "NEWS_SENTIMENT")
	q.Set("apikey", c.alphaVantageAPIKey)
	if len(targets) > 0 {
		tickers := make([]string, 0, len(targets))
		for t := range targets {
			// avoid alias words for AV ticker query
			if t == "BITCOIN" || t == "ETHEREUM" || t == "SOLANA" || t == "RIPPLE" {
				continue
			}
			tickers = append(tickers, t)
		}
		sort.Strings(tickers)
		if len(tickers) > 0 {
			q.Set("tickers", strings.Join(tickers, ","))
		}
	}
	endpoint.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	req.Header.Set("User-Agent", "NOFX-NewsFetcher/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("alphavantage status %d", resp.StatusCode)
	}

	var payload struct {
		Feed []struct {
			Title                 string `json:"title"`
			URL                   string `json:"url"`
			Source                string `json:"source"`
			TimePublished         string `json:"time_published"`
			OverallSentimentScore string `json:"overall_sentiment_score"`
			TickerSentiment       []struct {
				Ticker         string `json:"ticker"`
				RelevanceScore string `json:"relevance_score"`
			} `json:"ticker_sentiment"`
		} `json:"feed"`
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 3*1024*1024))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	result := make([]Item, 0, len(payload.Feed))
	for _, it := range payload.Feed {
		t, ok := parseAVTime(it.TimePublished)
		if !ok || t.Before(since) {
			continue
		}
		title := strings.TrimSpace(it.Title)
		if title == "" {
			continue
		}

		syms := make([]string, 0, len(it.TickerSentiment))
		for _, ts := range it.TickerSentiment {
			tk := strings.ToUpper(strings.TrimSpace(ts.Ticker))
			if tk != "" {
				syms = append(syms, tk)
			}
		}
		if len(syms) == 0 {
			syms = matchSymbols(title, targets)
		}
		if len(syms) == 0 && !isRelevantHeadline(title) {
			continue
		}

		sent := sentimentScore(title)
		if c.useAlphaVantageSentiment {
			if v, err := strconv.ParseFloat(strings.TrimSpace(it.OverallSentimentScore), 64); err == nil {
				// AV score often already in [-1,1], clamp just in case
				if v > 1 {
					v = 1
				} else if v < -1 {
					v = -1
				}
				sent = v
			}
		}

		source := strings.TrimSpace(it.Source)
		if source == "" {
			source = "Alpha Vantage"
		}
		result = append(result, Item{
			Title:       title,
			Link:        strings.TrimSpace(it.URL),
			Source:      source,
			PublishedAt: t,
			Symbols:     syms,
			Sentiment:   sent,
		})
	}

	return result, nil
}

func parseAVTime(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	layouts := []string{"20060102T150405", "20060102T1504", "20060102T150405Z"}
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t.UTC(), true
		}
	}
	return parseTime(v)
}

func parseTime(v string) (time.Time, bool) {
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC822Z, time.RFC822}
	v = strings.TrimSpace(v)
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func normalizeTargets(symbols []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range symbols {
		s = strings.ToUpper(strings.TrimSpace(s))
		if strings.HasSuffix(s, "USDT") {
			s = strings.TrimSuffix(s, "USDT")
		}
		if s != "" {
			m[s] = true
		}
	}
	// major aliases
	if m["BTC"] { m["BITCOIN"] = true }
	if m["ETH"] { m["ETHEREUM"] = true }
	if m["SOL"] { m["SOLANA"] = true }
	if m["XRP"] { m["RIPPLE"] = true }
	return m
}

func matchSymbols(title string, targets map[string]bool) []string {
	up := strings.ToUpper(title)
	hits := make([]string, 0, 3)
	for token := range targets {
		if strings.Contains(up, token) {
			hits = append(hits, token)
		}
	}
	sort.Strings(hits)
	if len(hits) > 3 {
		return hits[:3]
	}
	return hits
}

func sentimentScore(text string) float64 {
	up := strings.ToUpper(text)
	pos := []string{"SURGE", "RALLY", "BREAKOUT", "APPROVAL", "BULL", "GAIN", "UPGRADE"}
	neg := []string{"HACK", "SEC", "LAWSUIT", "BAN", "CRASH", "SELL-OFF", "DOWNGRADE", "EXPLOIT"}
	score := 0.0
	for _, k := range pos { if strings.Contains(up, k) { score += 1 } }
	for _, k := range neg { if strings.Contains(up, k) { score -= 1 } }
	if score > 3 { score = 3 }
	if score < -3 { score = -3 }
	return score / 3.0
}

func isRelevantHeadline(title string) bool {
	up := strings.ToUpper(strings.TrimSpace(title))
	if up == "" {
		return false
	}
	keywords := []string{
		"CRYPTO", "BITCOIN", "ETHEREUM", "ALTCOIN", "BLOCKCHAIN", "TOKEN",
		"ETF", "FED", "SEC", "CPI", "FOMC", "RATE CUT", "RATE HIKE",
		"BINANCE", "COINBASE", "HACK", "REGULATION", "LIQUIDATION",
	}
	for _, k := range keywords {
		if strings.Contains(up, k) {
			return true
		}
	}
	return false
}

func sourceHost(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	parts := strings.Split(u, "/")
	if len(parts) > 0 { return parts[0] }
	return u
}

func min(a, b int) int { if a < b { return a }; return b }
