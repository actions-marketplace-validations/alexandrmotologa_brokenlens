package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/crawler"
	"github.com/alexandrmotologa/brokenlens/internal/model"
	"github.com/alexandrmotologa/brokenlens/internal/reporter"
	"github.com/alexandrmotologa/brokenlens/internal/server"
)

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Standalone server mode from report file
	if cfg.Serve && cfg.TargetURL == "" && cfg.ReportFile != "" {
		srv := server.NewServer(cfg, nil)
		if err := srv.LoadReportFromFile(cfg.ReportFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load report from %s: %v\n", cfg.ReportFile, err)
			os.Exit(1)
		}
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		if err := srv.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// If serve mode without target URL and without report, open blank studio
	if cfg.Serve && cfg.TargetURL == "" {
		srv := server.NewServer(cfg, nil)
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		if err := srv.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Context for graceful shutdown on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nReceived termination signal. Finishing in-flight requests...")
		cancel()
	}()

	termReporter := reporter.NewTerminalReporter(os.Stdout, cfg.Quiet)

	engine, err := crawler.NewEngine(cfg, termReporter.OnProgress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Quiet {
		fmt.Printf("[brokenlens] Starting high-speed audit of %s (concurrency: %d, max-depth: %d)\n",
			cfg.TargetURL, cfg.Concurrency, cfg.MaxDepth)
	}

	report, err := engine.Run(ctx)
	if err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Crawl execution error: %v\n", err)
		os.Exit(1)
	}

	if report == nil {
		fmt.Fprintln(os.Stderr, "No crawl results produced.")
		os.Exit(1)
	}

	// Render CLI terminal summary
	if !cfg.Quiet {
		termReporter.RenderSummary(report)
	}

	// Handle File Exports
	if cfg.OutputFile != "" {
		switch cfg.OutputFormat {
		case "csv":
			if err := reporter.ExportCSV(cfg.OutputFile, report.Issues); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to export CSV: %v\n", err)
			} else if !cfg.Quiet {
				fmt.Printf("[brokenlens] Issues exported to CSV: %s\n", cfg.OutputFile)
			}
		default: // json
			if err := reporter.ExportJSON(cfg.OutputFile, report); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to export JSON: %v\n", err)
			} else if !cfg.Quiet {
				fmt.Printf("[brokenlens] Report exported to JSON: %s\n", cfg.OutputFile)
			}
		}
	}

	if cfg.JUnitFile != "" {
		if err := reporter.ExportJUnit(cfg.JUnitFile, report); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export JUnit XML: %v\n", err)
		} else if !cfg.Quiet {
			fmt.Printf("[brokenlens] JUnit XML exported to: %s\n", cfg.JUnitFile)
		}
	}

	if cfg.MarkdownFile != "" {
		if err := reporter.ExportMarkdown(cfg.MarkdownFile, report); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export Markdown summary: %v\n", err)
		} else if !cfg.Quiet {
			fmt.Printf("[brokenlens] Markdown summary exported to: %s\n", cfg.MarkdownFile)
		}
	}

	// GitHub Actions Step Summary auto-detection
	_ = reporter.WriteGitHubStepSummary(report)

	// GitHub Actions Step Output auto-detection
	if ghOutput := os.Getenv("GITHUB_OUTPUT"); ghOutput != "" {
		if f, err := os.OpenFile(ghOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "total-scanned=%d\n", report.Summary.TotalScanned)
			fmt.Fprintf(f, "total-broken=%d\n", report.Summary.TotalBroken)
			fmt.Fprintf(f, "health-score=%.2f\n", report.Summary.HealthScore)
			fmt.Fprintf(f, "total-redirects=%d\n", report.Summary.TotalRedirects)
			fmt.Fprintf(f, "total-mixed-content=%d\n", report.Summary.TotalMixedContent)
			if cfg.OutputFile != "" {
				fmt.Fprintf(f, "report-path=%s\n", cfg.OutputFile)
			}
			if cfg.JUnitFile != "" {
				fmt.Fprintf(f, "junit-path=%s\n", cfg.JUnitFile)
			}
			_ = f.Close()
		}
	}

	// If --serve was requested alongside TargetURL, start server with completed report
	if cfg.Serve {
		srv := server.NewServer(cfg, report)
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		if err := srv.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CI/CD Gating Verification
	if cfg.FailOnBroken {
		criticalCount := 0
		for _, issue := range report.Issues {
			if issue.Severity == model.SeverityCritical {
				criticalCount++
			}
		}

		if report.Summary.TotalBroken > cfg.MaxBrokenThreshold || criticalCount > 0 {
			fmt.Fprintf(os.Stderr, "\n[brokenlens] CI/CD Gating Failure: Found %d broken items and %d critical issues (allowed threshold: %d).\n",
				report.Summary.TotalBroken, criticalCount, cfg.MaxBrokenThreshold)
			os.Exit(1)
		}
	}

	// CI/CD Differential PR Gating: Fail only on NEW broken issues
	if cfg.FailOnNewBroken {
		if report.Diff == nil {
			fmt.Fprintln(os.Stderr, "\n[brokenlens] Error: --fail-on-new-broken requires a baseline report specified via --diff <path>")
			os.Exit(2)
		}
		if len(report.Diff.NewIssues) > 0 {
			fmt.Fprintf(os.Stderr, "\n[brokenlens] CI/CD Differential Gating Failure: %d NEW broken issue(s) introduced relative to baseline %s.\n",
				len(report.Diff.NewIssues), cfg.DiffBaseline)
			for _, iss := range report.Diff.NewIssues {
				fmt.Fprintf(os.Stderr, "  - [%s] %s (%s) on %s\n", iss.Severity, iss.TargetURL, iss.Type, iss.SourceURL)
			}
			os.Exit(1)
		}
	}
}
