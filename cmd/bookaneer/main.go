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
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
	"github.com/woliveiras/bookaneer/internal/metadata"
	"github.com/woliveiras/bookaneer/internal/metadata/googlebooks"
	"github.com/woliveiras/bookaneer/internal/metadata/openlibrary"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Load .env file if present (optional, won't error if missing)
	_ = godotenv.Load()

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

	// Create default admin user if no users exist
	// Supports BOOKANEER_ADMIN_PASSWORD env var for Docker deployments
	envPassword := os.Getenv("BOOKANEER_ADMIN_PASSWORD")
	adminPassword, err := authSvc.EnsureDefaultAdmin(context.Background(), envPassword)
	if err != nil {
		return fmt.Errorf("ensure default admin: %w", err)
	}
	if adminPassword != "" {
		slog.Info("===========================================")
		slog.Info("Default admin user created")
		slog.Info("Username: admin")
		if envPassword != "" {
			slog.Info("Password: <set via BOOKANEER_ADMIN_PASSWORD>")
		} else {
			slog.Info("Password: " + adminPassword)
			// Save to file for Docker users who miss the log
			credentialsFile := filepath.Join(cfg.DataDir, "admin_credentials.txt")
			content := fmt.Sprintf("Bookaneer Default Admin Credentials\n\nUsername: admin\nPassword: %s\n\nDelete this file after you have saved these credentials.\n", adminPassword)
			if err := os.WriteFile(credentialsFile, []byte(content), 0600); err != nil {
				slog.Warn("could not save credentials file", "error", err)
			} else {
				slog.Info("Credentials also saved to: " + credentialsFile)
			}
		}
		slog.Info("Please change your password after first login!")
		slog.Info("===========================================")
	}

	// Log the API key for external integrations
	apiKey, err := authSvc.GetAPIKey(context.Background())
	if err == nil && apiKey != "" {
		slog.Info("API key for external integrations (OPDS, scripts, etc)", "apiKey", apiKey)
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

	// Settings handler (protected - shows API key)
	settingsHandler := handler.NewSettingsHandler(authSvc, cfg)
	settingsHandler.Register(protected)

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

	// Metadata providers (OpenLibrary + GoogleBooks)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	olProvider := openlibrary.New(httpClient, "Bookaneer/1.0 (https://github.com/woliveiras/bookaneer)")
	gbProvider := googlebooks.New(httpClient, "") // No API key needed for basic usage

	metaAggregator := metadata.NewAggregator(slog.Default(), olProvider, gbProvider)
	metadataHandler := handler.NewMetadataHandler(metaAggregator)
	protected.GET("/metadata/authors", metadataHandler.SearchAuthors)
	protected.GET("/metadata/books", metadataHandler.SearchBooks)
	protected.GET("/metadata/authors/:foreignId", metadataHandler.GetAuthor)
	protected.GET("/metadata/books/:foreignId", metadataHandler.GetBook)
	protected.GET("/metadata/isbn/:isbn", metadataHandler.LookupISBN)
	protected.GET("/metadata/providers", metadataHandler.ListProviders)

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
