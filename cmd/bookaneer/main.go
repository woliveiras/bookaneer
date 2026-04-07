package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	bookaneer "github.com/woliveiras/bookaneer"
	"github.com/woliveiras/bookaneer/api/v1/handler"
	apimw "github.com/woliveiras/bookaneer/api/v1/middleware"
	"github.com/woliveiras/bookaneer/internal/auth"
	"github.com/woliveiras/bookaneer/internal/config"
	"github.com/woliveiras/bookaneer/internal/core/author"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/library"
	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
	"github.com/woliveiras/bookaneer/internal/core/rootfolder"
	"github.com/woliveiras/bookaneer/internal/core/series"
	"github.com/woliveiras/bookaneer/internal/database"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath = flag.String("config", "", "path to config.yaml")
		dataDir    = flag.String("data", "", "data directory")
		showVer    = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("bookaneer %s (built %s)\n", version, buildTime)
		return nil
	}

	cfg, err := config.Load(*dataDir, *configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	setupLogging(cfg.LogLevel)
	slog.Info("starting bookaneer", "version", version, "dataDir", cfg.DataDir)

	db, err := database.Open(cfg.DatabasePath())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db, bookaneer.MigrationsFS, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	authSvc := auth.New(db)
	if err := authSvc.EnsureAPIKey(context.Background()); err != nil {
		return fmt.Errorf("ensure api key: %w", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(apimw.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-Api-Key"},
	}))

	api := e.Group("/api/v1")

	systemHandler := handler.NewSystemHandler(version, buildTime, cfg)
	api.GET("/system/status", systemHandler.Status)
	api.GET("/system/health", systemHandler.Health)

	protected := api.Group("")
	protected.Use(apimw.Auth(authSvc))

	authHandler := handler.NewAuthHandler(authSvc)
	api.POST("/auth/login", authHandler.Login)
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)

	// Core domain services
	authorSvc := author.New(db)
	bookSvc := book.New(db)
	seriesSvc := series.New(db)
	rootFolderSvc := rootfolder.New(db)
	qualityProfileSvc := qualityprofile.New(db)
	libraryScanner := library.NewScanner(db)

	// Ensure default quality profile exists
	if err := qualityProfileSvc.EnsureDefault(context.Background()); err != nil {
		slog.Warn("could not ensure default quality profile", "error", err)
	}

	// Core domain handlers
	authorHandler := handler.NewAuthorHandler(authorSvc)
	authorHandler.Register(protected)

	bookHandler := handler.NewBookHandler(bookSvc)
	bookHandler.Register(protected)

	seriesHandler := handler.NewSeriesHandler(seriesSvc)
	seriesHandler.Register(protected)

	rootFolderHandler := handler.NewRootFolderHandler(rootFolderSvc)
	rootFolderHandler.Register(protected)

	qualityProfileHandler := handler.NewQualityProfileHandler(qualityProfileSvc)
	qualityProfileHandler.Register(protected)

	libraryHandler := handler.NewLibraryHandler(libraryScanner)
	libraryHandler.Register(protected)

	if err := serveFrontend(e); err != nil {
		return fmt.Errorf("setup frontend: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)
	go func() {
		slog.Info("listening", "address", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("goodbye")
	return nil
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
}

func serveFrontend(e *echo.Echo) error {
	distFS, err := fs.Sub(bookaneer.WebFS, "web/dist")
	if err != nil {
		slog.Warn("embedded frontend not found, serving from filesystem")
		e.Static("/", "web/dist")
		return nil
	}

	fileServer := http.FileServer(http.FS(distFS))
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if f, err := distFS.Open(path[1:]); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})))

	return nil
}
