// Package main provides the entry point for the Yandex MCP server.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/n-r-w/yandex-mcp/internal/adapters/authremote"
	"github.com/n-r-w/yandex-mcp/internal/adapters/tracker"
	"github.com/n-r-w/yandex-mcp/internal/adapters/wiki"
	"github.com/n-r-w/yandex-mcp/internal/adapters/yc"
	"github.com/n-r-w/yandex-mcp/internal/adapters/ytoken"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
	trackertools "github.com/n-r-w/yandex-mcp/internal/tools/tracker"
	wikitools "github.com/n-r-w/yandex-mcp/internal/tools/wiki"
)

// build-time variables that can be set via ldflags
//
//nolint:gochecknoglobals // global variables are used for build-time information
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
	flag.Parse()

	info := getBuildInfo()

	if *showVersion {
		//nolint:exhaustruct_v5 // stdlib struct with optional fields
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		logger.Info(
			"yandex-mcp version info",
			"version", info.version,
			"commit", info.commit,
			"built", info.date,
			"built_by", info.builtBy,
		)
		os.Exit(0)
	}

	//nolint:exhaustruct_v5 // SDK struct with optional fields
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(info.version, flag.Args()); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(serverVersion string, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if len(args) > 0 {
		if len(args) != 1 || args[0] != "auth-agent" {
			return errors.New("usage: yandex-mcp [-version] [auth-agent]")
		}
		cfg, err := authagent.LoadConfig()
		if err != nil {
			return err
		}
		service := authagent.New(yc.New(cfg.YCPath), cfg.Profiles)
		defer service.Close()
		slog.InfoContext(ctx, "starting workstation auth-agent", "address", cfg.ListenAddress)
		return service.Run(ctx, cfg.ListenAddress)
	}
	return runMCP(ctx, serverVersion)
}

func runMCP(ctx context.Context, serverVersion string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	slog.InfoContext(ctx,
		"configuration loaded",
		slog.String("wiki_base_url", cfg.WikiBaseURL),
		slog.String("tracker_base_url", cfg.TrackerBaseURL),
	)

	var tokenProvider *ytoken.Provider
	if cfg.TokenSource == "remote" {
		source, sourceErr := authremote.New(cfg.AuthAgentURL)
		if sourceErr != nil {
			return sourceErr
		}
		defer source.Close()
		tokenProvider = ytoken.New(source, cfg.CLIProfile, cfg.IAMTokenRefreshPeriod)
	} else {
		tokenProvider = ytoken.New(yc.New("yc"), cfg.CLIProfile, cfg.IAMTokenRefreshPeriod)
	}
	defer tokenProvider.Close()

	wikiClient := wiki.NewClient(cfg, tokenProvider)
	trackerClient := tracker.NewClient(cfg, tokenProvider)

	wikiTools := domain.WikiAllTools()
	trackerTools := domain.TrackerAllTools()

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiClient, wikiTools),
		trackertools.NewRegistrator(
			trackerClient,
			trackerTools,
			cfg.AttachAllowedExtensions,
			cfg.AttachViewExtensions,
			cfg.AttachAllowedDirs,
		),
	}

	srv, err := server.New(serverVersion, registrators, cfg.ToolTimeout)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "starting MCP server over stdio")

	transport := &mcp.StdioTransport{}
	return srv.Run(ctx, transport)
}
