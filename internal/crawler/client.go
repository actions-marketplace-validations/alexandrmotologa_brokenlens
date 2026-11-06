package crawler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// FetchResponse wraps the HTTP response details and content.
type FetchResponse struct {
	StatusCode     int
	StatusText     string
	DurationMs     int64
	ContentType    string
	ContentLength  int64
	FinalURL       string
	RedirectChain  []model.RedirectHop
	IsRedirectLoop bool
	XRobotsTag     string
	Body           []byte
	Error          error
}

// HTTPClient manages outbound HTTP requests with connection pooling and redirect tracking.
type HTTPClient struct {
	client       *http.Client
	cfg          *config.Config
	transport    *http.Transport
}

// NewHTTPClient creates an optimized HTTP client for high-throughput crawling.
func NewHTTPClient(cfg *config.Config) *HTTPClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          500,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &HTTPClient{
		cfg:       cfg,
		transport: transport,
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
			// Do not follow redirects automatically; we follow them manually to record each hop
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Fetch executes an HTTP request, manually following redirects to record all hops.
func (c *HTTPClient) Fetch(ctx context.Context, targetURL string, method string, readBody bool) *FetchResponse {
	start := time.Now()
	currentURL := targetURL
	var chain []model.RedirectHop
	seenURLs := make(map[string]bool)
	seenURLs[targetURL] = true

	for hopCount := 0; hopCount <= c.cfg.MaxRedirects; hopCount++ {
		req, err := http.NewRequestWithContext(ctx, method, currentURL, nil)
		if err != nil {
			return &FetchResponse{
				StatusCode: 0,
				DurationMs: time.Since(start).Milliseconds(),
				FinalURL:   currentURL,
				Error:      fmt.Errorf("create request failed: %w", err),
			}
		}

		req.Header.Set("User-Agent", c.cfg.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		// Custom Headers
		for k, v := range c.cfg.Headers {
			req.Header.Set(k, v)
		}

		// Cookie
		if c.cfg.Cookie != "" {
			req.Header.Set("Cookie", c.cfg.Cookie)
		}

		// Basic Auth
		if c.cfg.BasicAuth != "" {
			parts := strings.SplitN(c.cfg.BasicAuth, ":", 2)
			if len(parts) == 2 {
				req.SetBasicAuth(parts[0], parts[1])
			}
		}

		resp, err := c.client.Do(req)
		if err != nil {
			// External servers might reject HEAD; fallback to GET if method was HEAD
			if method == http.MethodHead {
				return c.Fetch(ctx, targetURL, http.MethodGet, false)
			}
			return &FetchResponse{
				StatusCode:    0,
				DurationMs:    time.Since(start).Milliseconds(),
				FinalURL:      currentURL,
				RedirectChain: chain,
				Error:         err,
			}
		}

		statusCode := resp.StatusCode

		// Smart retry on 429 Too Many Requests or 503 with Retry-After header
		if statusCode == http.StatusTooManyRequests || statusCode == http.StatusServiceUnavailable {
			retryAfterStr := resp.Header.Get("Retry-After")
			_ = resp.Body.Close()
			if retryAfterStr != "" {
				waitSec := 1
				if s, err := strconv.Atoi(retryAfterStr); err == nil && s > 0 {
					if s > 3 {
						waitSec = 3 // Cap retry backoff to 3 seconds
					} else {
						waitSec = s
					}
				}
				time.Sleep(time.Duration(waitSec) * time.Second)
				// Re-attempt once
				resp, err = c.client.Do(req)
				if err == nil {
					statusCode = resp.StatusCode
				}
			}
		}

		statusText := http.StatusText(statusCode)
		contentType := resp.Header.Get("Content-Type")
		contentLength := resp.ContentLength
		xRobotsTag := resp.Header.Get("X-Robots-Tag")

		// Handle Redirection (301, 302, 303, 307, 308)
		if isRedirect(statusCode) {
			location := resp.Header.Get("Location")
			_ = resp.Body.Close()

			if location == "" {
				return &FetchResponse{
					StatusCode:    statusCode,
					StatusText:    statusText,
					DurationMs:    time.Since(start).Milliseconds(),
					FinalURL:      currentURL,
					RedirectChain: chain,
					Error:         errors.New("redirect status without Location header"),
				}
			}

			// Resolve relative redirect location against current URL
			nextURL, err := resolveURL(currentURL, location)
			if err != nil {
				return &FetchResponse{
					StatusCode:    statusCode,
					StatusText:    statusText,
					DurationMs:    time.Since(start).Milliseconds(),
					FinalURL:      currentURL,
					RedirectChain: chain,
					Error:         fmt.Errorf("invalid redirect location %q: %w", location, err),
				}
			}

			chain = append(chain, model.RedirectHop{
				FromURL:    currentURL,
				ToURL:      nextURL,
				StatusCode: statusCode,
			})

			// Check for circular loop
			if seenURLs[nextURL] {
				return &FetchResponse{
					StatusCode:     statusCode,
					StatusText:     statusText,
					DurationMs:     time.Since(start).Milliseconds(),
					FinalURL:       nextURL,
					RedirectChain:  chain,
					IsRedirectLoop: true,
					Error:          fmt.Errorf("circular redirect loop detected: %s -> %s", currentURL, nextURL),
				}
			}
			seenURLs[nextURL] = true
			currentURL = nextURL

			// If method was HEAD and we redirect, continue with HEAD
			continue
		}

		// Non-redirect response reached:
		var bodyBytes []byte
		if readBody && method != http.MethodHead {
			// Limit body read to 10MB to prevent memory exhaustion on giant media
			bodyBytes, err = io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
			_ = resp.Body.Close()
			if err != nil {
				return &FetchResponse{
					StatusCode:    statusCode,
					StatusText:    statusText,
					DurationMs:    time.Since(start).Milliseconds(),
					FinalURL:      currentURL,
					RedirectChain: chain,
					ContentType:   contentType,
					ContentLength: contentLength,
					Error:         fmt.Errorf("reading body failed: %w", err),
				}
			}
		} else {
			_ = resp.Body.Close()
		}

		// If HEAD returned 405 Method Not Allowed or 403 Forbidden, fallback to GET without body reading
		if method == http.MethodHead && (statusCode == http.StatusMethodNotAllowed || statusCode == http.StatusForbidden) {
			return c.Fetch(ctx, targetURL, http.MethodGet, false)
		}

		return &FetchResponse{
			StatusCode:    statusCode,
			StatusText:    statusText,
			DurationMs:    time.Since(start).Milliseconds(),
			ContentType:   contentType,
			ContentLength: contentLength,
			FinalURL:      currentURL,
			RedirectChain: chain,
			XRobotsTag:    xRobotsTag,
			Body:          bodyBytes,
		}
	}

	return &FetchResponse{
		StatusCode:    310,
		StatusText:    "Too Many Redirects",
		DurationMs:    time.Since(start).Milliseconds(),
		FinalURL:      currentURL,
		RedirectChain: chain,
		Error:         fmt.Errorf("exceeded maximum redirect limit of %d hops", c.cfg.MaxRedirects),
	}
}

func isRedirect(status int) bool {
	return status == http.StatusMovedPermanently ||
		status == http.StatusFound ||
		status == http.StatusSeeOther ||
		status == http.StatusTemporaryRedirect ||
		status == http.StatusPermanentRedirect
}

func resolveURL(base, ref string) (string, error) {
	baseURL, err := http.NewRequest("GET", base, nil)
	if err != nil {
		return "", err
	}
	relURL, err := baseURL.URL.Parse(ref)
	if err != nil {
		return "", err
	}
	return relURL.String(), nil
}

// IsHTMLContent checks if the Content-Type header represents an HTML document.
func IsHTMLContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml")
}
