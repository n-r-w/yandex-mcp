// Package main provides the entry point for the Yandex MCP server.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n-r-w/yandex-mcp/internal/adapters/token"
	"github.com/n-r-w/yandex-mcp/internal/adapters/tracker"
	"github.com/n-r-w/yandex-mcp/internal/adapters/wiki"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
	trackertools "github.com/n-r-w/yandex-mcp/internal/tools/tracker"
	wikitools "github.com/n-r-w/yandex-mcp/internal/tools/wiki"
)

// build-time variables that can be set via ldflags
//
//nolint:nolintlint // gochecknoglobals is excluded for this file via .golangci.yml
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

// buildInfo holds build-time information.
type buildInfo struct {
	version string
	commit  string
	date    string
	builtBy string
}

// getBuildInfo returns build-time information.
func getBuildInfo() buildInfo {
	return buildInfo{
		version: version,
		commit:  commit,
		date:    date,
		builtBy: builtBy,
	}
}

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	wikiWrite := flag.Bool("wiki-write", false, "enable write operations for Yandex Wiki tools")
	trackerWrite := flag.Bool("tracker-write", false, "enable write operations for Yandex Tracker tools")
	flag.Parse()

	info := getBuildInfo()

	if *showVersion {
		//nolint:exhaustruct // stdlib struct with optional fields
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		logger.Info("yandex-mcp version info",
			"version", info.version,
			"commit", info.commit,
			"built", info.date,
			"built_by", info.builtBy,
		)
		os.Exit(0)
	}

	//nolint:exhaustruct // SDK struct with optional fields
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(info.version, *wikiWrite, *trackerWrite); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(serverVersion string, wikiWrite, trackerWrite bool) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	slog.Info("configuration loaded",
		slog.String("wiki_base_url", cfg.WikiBaseURL),
		slog.String("tracker_base_url", cfg.TrackerBaseURL),
	)

	tokenProvider := token.NewProvider(cfg)

	wikiClient := wiki.NewClient(cfg, tokenProvider)
	trackerClient := tracker.NewClient(cfg, tokenProvider)

	wikiTools := buildWikiToolList(wikiWrite)
	trackerTools := buildTrackerToolList(trackerWrite)

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiClient, wikiTools),
		trackertools.NewRegistrator(trackerClient, trackerTools),
	}

	srv, err := server.New(serverVersion, registrators)
	if err != nil {
		return err
	}

	slog.Info("starting MCP server over stdio",
		slog.Bool("wiki_write", wikiWrite),
		slog.Bool("tracker_write", trackerWrite),
	)

	transport := &mcp.StdioTransport{}
	return srv.Run(ctx, transport)
}

// buildWikiToolList returns the list of wiki tools based on flags.
func buildWikiToolList(wikiWrite bool) []domain.WikiTool {
	if wikiWrite {
		return domain.WikiAllTools()
	}
	return domain.WikiReadOnlyTools()
}

// buildTrackerToolList returns the list of tracker tools based on flags.
func buildTrackerToolList(trackerWrite bool) []domain.TrackerTool {
	if trackerWrite {
		return domain.TrackerAllTools()
	}
	return domain.TrackerReadOnlyTools()
}
