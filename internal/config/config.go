package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Config holds runtime configuration options for the auditor.
type Config struct {
	TargetURL          string
	Concurrency        int
	MaxDepth           int
	MaxURLs            int
	Timeout            time.Duration
	UserAgent          string
	CheckExternal      bool
	CheckFragments     bool
	CheckMixedContent  bool
	FollowRobotsTxt    bool
	HonorSitemap       bool
	MaxRedirects       int
	FailOnBroken       bool
	MaxBrokenThreshold int
	OutputFile         string
	OutputFormat       string
	JUnitFile          string
	MarkdownFile       string
	Serve              bool
	Host               string
	Port               int
	IgnoreHosts        []string
	IncludePattern     string
	ExcludePattern     string
	RateLimitMs        int
	Quiet              bool
	ReportFile         string // For 'brokenlens serve --report report.json'

	// Auth and custom request parameters
	BasicAuth       string
	Headers         map[string]string
	Cookie          string

	// Differential comparison and advanced SEO
	DiffBaseline    string
	FailOnNewBroken bool
	CheckSEO        bool

	// Compiled filters
	IncludeRegex *regexp.Regexp
	ExcludeRegex *regexp.Regexp
}

// DefaultConfig returns the standard default configuration values.
func DefaultConfig() *Config {
	return &Config{
		Concurrency:        25,
		MaxDepth:           10,
		MaxURLs:            0,
		Timeout:            10 * time.Second,
		UserAgent:          "BrokenLens/1.0 (+https://github.com/alexandrmotologa/brokenlens)",
		CheckExternal:      true,
		CheckFragments:     true,
		CheckMixedContent:  true,
		FollowRobotsTxt:    true,
		HonorSitemap:       true,
		MaxRedirects:       10,
		FailOnBroken:       false,
		MaxBrokenThreshold: 0,
		Host:               "127.0.0.1",
		Port:               8080,
		RateLimitMs:        0,
		Quiet:              false,
		Headers:            make(map[string]string),
		CheckSEO:           true,
	}
}

// ParseFlags parses command line arguments and environment variables into Config.
func ParseFlags(args []string) (*Config, error) {
	cfg := DefaultConfig()

	fs := flag.NewFlagSet("brokenlens", flag.ContinueOnError)

	var ignoreHostsStr string
	var headersStr string
	var timeoutSec int

	fs.StringVar(&cfg.BasicAuth, "auth", "", "HTTP Basic Auth credentials in username:password format")
	fs.StringVar(&headersStr, "header", "", "Custom HTTP header (key:value or multiple separated by ||)")
	fs.StringVar(&cfg.Cookie, "cookie", "", "Custom Cookie header string")
	fs.StringVar(&cfg.DiffBaseline, "diff", "", "Path to baseline report JSON file for differential comparison")
	fs.BoolVar(&cfg.FailOnNewBroken, "fail-on-new-broken", false, "Exit with code 1 only if new broken items appear relative to baseline")
	fs.BoolVar(&cfg.CheckSEO, "check-seo", cfg.CheckSEO, "Audit canonical tags, meta robots, social images, and soft 404s")

	fs.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "Number of concurrent worker routines (1-250)")
	fs.IntVar(&cfg.Concurrency, "c", cfg.Concurrency, "Shorthand for --concurrency")

	fs.IntVar(&cfg.MaxDepth, "max-depth", cfg.MaxDepth, "Maximum link crawl depth (0 = unlimited)")
	fs.IntVar(&cfg.MaxDepth, "d", cfg.MaxDepth, "Shorthand for --max-depth")

	fs.IntVar(&cfg.MaxURLs, "max-urls", cfg.MaxURLs, "Maximum number of URLs to crawl (0 = unlimited)")
	fs.IntVar(&timeoutSec, "timeout", 10, "Per-request timeout in seconds")

	fs.StringVar(&cfg.UserAgent, "user-agent", cfg.UserAgent, "Custom User-Agent header")
	fs.BoolVar(&cfg.CheckExternal, "check-external", cfg.CheckExternal, "Verify HTTP status of external links")
	fs.BoolVar(&cfg.CheckFragments, "check-fragments", cfg.CheckFragments, "Verify in-page #anchor existence in target DOM")
	fs.BoolVar(&cfg.CheckMixedContent, "check-mixed-content", cfg.CheckMixedContent, "Detect HTTP assets on HTTPS pages")
	fs.BoolVar(&cfg.FollowRobotsTxt, "robots", cfg.FollowRobotsTxt, "Respect robots.txt rules")
	fs.BoolVar(&cfg.HonorSitemap, "sitemap", cfg.HonorSitemap, "Auto-discover and crawl URLs from sitemap.xml")

	fs.IntVar(&cfg.MaxRedirects, "max-redirects", cfg.MaxRedirects, "Maximum redirect hops before reporting failure")
	fs.BoolVar(&cfg.FailOnBroken, "fail-on-broken", cfg.FailOnBroken, "Exit with code 1 if broken links or critical issues are found")
	fs.IntVar(&cfg.MaxBrokenThreshold, "max-broken", cfg.MaxBrokenThreshold, "Maximum allowed broken items before exiting with failure")

	fs.StringVar(&cfg.OutputFile, "output", "", "Export report file path (.json, .csv)")
	fs.StringVar(&cfg.OutputFile, "o", "", "Shorthand for --output")
	fs.StringVar(&cfg.OutputFormat, "format", "", "Explicit export format (json, csv, junit, markdown)")
	fs.StringVar(&cfg.JUnitFile, "junit", "", "Export JUnit XML report file path")
	fs.StringVar(&cfg.MarkdownFile, "markdown", "", "Export Markdown step summary file path")

	fs.BoolVar(&cfg.Serve, "serve", false, "Start embedded React visual report studio")
	fs.StringVar(&cfg.Host, "host", cfg.Host, "Host address for the report server")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "Port for the report server")
	fs.StringVar(&cfg.ReportFile, "report", "", "Load existing report JSON file for visual server viewer")

	fs.StringVar(&ignoreHostsStr, "ignore-hosts", "", "Comma-separated list of hostnames to ignore")
	fs.StringVar(&cfg.IncludePattern, "include", "", "Regex pattern of URLs to include in crawl")
	fs.StringVar(&cfg.ExcludePattern, "exclude", "", "Regex pattern of URLs to exclude from crawl")
	fs.IntVar(&cfg.RateLimitMs, "delay", 0, "Delay in milliseconds between requests per worker")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "Suppress terminal live spinner and non-error output")
	fs.BoolVar(&cfg.Quiet, "q", false, "Shorthand for --quiet")

	// Pre-process args to separate positional arguments from flags,
	// allowing flags anywhere (before or after the target URL)
	valueFlags := map[string]bool{
		"concurrency": true, "c": true,
		"max-depth": true, "d": true,
		"max-urls": true, "timeout": true,
		"user-agent": true, "max-redirects": true,
		"max-broken": true, "output": true, "o": true,
		"format": true, "junit": true, "markdown": true,
		"host": true, "port": true, "report": true,
		"ignore-hosts": true, "include": true, "exclude": true,
		"delay": true, "auth": true, "header": true, "cookie": true, "diff": true,
	}

	var flagArgs []string
	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			flagName := strings.TrimLeft(arg, "-")
			// Handle --flag=val format
			if strings.Contains(flagName, "=") {
				continue
			}
			// If this flag requires a following value argument
			if valueFlags[flagName] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if err := fs.Parse(flagArgs); err != nil {
		return nil, err
	}

	cfg.Timeout = time.Duration(timeoutSec) * time.Second

	// Process positional arguments
	if len(positionalArgs) > 0 {
		firstArg := positionalArgs[0]
		if firstArg == "serve" {
			cfg.Serve = true
			if len(positionalArgs) > 1 {
				cfg.TargetURL = positionalArgs[1]
			}
		} else {
			cfg.TargetURL = firstArg
		}
	}

	if ignoreHostsStr != "" {
		for _, h := range strings.Split(ignoreHostsStr, ",") {
			cleaned := strings.TrimSpace(h)
			if cleaned != "" {
				cfg.IgnoreHosts = append(cfg.IgnoreHosts, strings.ToLower(cleaned))
			}
		}
	}

	if headersStr != "" {
		for _, h := range strings.Split(headersStr, "||") {
			parts := strings.SplitN(strings.TrimSpace(h), ":", 2)
			if len(parts) == 2 {
				cfg.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	if cfg.DiffBaseline != "" {
		if _, err := os.Stat(cfg.DiffBaseline); err != nil {
			return nil, fmt.Errorf("diff baseline file not found: %s", cfg.DiffBaseline)
		}
	}

	// Validate settings
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks configuration values and compiles regular expressions.
func (c *Config) Validate() error {
	if c.Concurrency < 1 {
		c.Concurrency = 1
	}
	if c.Concurrency > 250 {
		c.Concurrency = 250
	}

	// If serve mode without TargetURL and ReportFile is provided, that's valid
	if c.Serve && c.TargetURL == "" && c.ReportFile != "" {
		if _, err := os.Stat(c.ReportFile); err != nil {
			return fmt.Errorf("report file not found: %s", c.ReportFile)
		}
		return nil
	}

	// Otherwise TargetURL is required
	if c.TargetURL == "" && !c.Serve {
		return errors.New("target URL is required (e.g. brokenlens https://example.com)")
	}

	if c.TargetURL != "" {
		if !strings.HasPrefix(c.TargetURL, "http://") && !strings.HasPrefix(c.TargetURL, "https://") {
			c.TargetURL = "https://" + c.TargetURL
		}
		parsed, err := url.Parse(c.TargetURL)
		if err != nil || parsed.Host == "" {
			return fmt.Errorf("invalid target URL: %s", c.TargetURL)
		}
	}

	if c.IncludePattern != "" {
		re, err := regexp.Compile(c.IncludePattern)
		if err != nil {
			return fmt.Errorf("invalid --include regex: %w", err)
		}
		c.IncludeRegex = re
	}

	if c.ExcludePattern != "" {
		re, err := regexp.Compile(c.ExcludePattern)
		if err != nil {
			return fmt.Errorf("invalid --exclude regex: %w", err)
		}
		c.ExcludeRegex = re
	}

	// Auto-detect format if not explicitly set
	if c.OutputFormat == "" && c.OutputFile != "" {
		lower := strings.ToLower(c.OutputFile)
		switch {
		case strings.HasSuffix(lower, ".json"):
			c.OutputFormat = "json"
		case strings.HasSuffix(lower, ".csv"):
			c.OutputFormat = "csv"
		case strings.HasSuffix(lower, ".xml"):
			c.OutputFormat = "junit"
		case strings.HasSuffix(lower, ".md"):
			c.OutputFormat = "markdown"
		default:
			c.OutputFormat = "json"
		}
	}

	return nil
}
