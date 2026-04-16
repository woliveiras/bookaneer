package main

import (
	"context"
	"database/sql"
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
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	bookaneer "github.com/woliveiras/bookaneer"
	"github.com/woliveiras/bookaneer/api/v1/handler"
	apimw "github.com/woliveiras/bookaneer/api/v1/middleware"
	"github.com/woliveiras/bookaneer/internal/auth"
	"github.com/woliveiras/bookaneer/internal/config"
	"github.com/woliveiras/bookaneer/internal/core/author"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/library"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/core/pathmapping"
	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
	"github.com/woliveiras/bookaneer/internal/core/reader"
	"github.com/woliveiras/bookaneer/internal/core/rootfolder"
	"github.com/woliveiras/bookaneer/internal/core/series"
	"github.com/woliveiras/bookaneer/internal/database"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/blackhole"
	_ "github.com/woliveiras/bookaneer/internal/download/direct"
	_ "github.com/woliveiras/bookaneer/internal/download/qbittorrent"
	_ "github.com/woliveiras/bookaneer/internal/download/sabnzbd"
	_ "github.com/woliveiras/bookaneer/internal/download/transmission"
	digitallibrary "github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/library/annas"
	"github.com/woliveiras/bookaneer/internal/library/aozora"
	"github.com/woliveiras/bookaneer/internal/library/archive"
	"github.com/woliveiras/bookaneer/internal/library/dominiopublico"
	"github.com/woliveiras/bookaneer/internal/library/gutendex"
	"github.com/woliveiras/bookaneer/internal/library/libgen"
	"github.com/woliveiras/bookaneer/internal/library/openlibrarypublic"
	"github.com/woliveiras/bookaneer/internal/library/sitesearch"
	"github.com/woliveiras/bookaneer/internal/library/wikisource"
	"github.com/woliveiras/bookaneer/internal/metadata"
	"github.com/woliveiras/bookaneer/internal/metadata/googlebooks"
	"github.com/woliveiras/bookaneer/internal/metadata/openlibrary"
	"github.com/woliveiras/bookaneer/internal/scheduler"
	"github.com/woliveiras/bookaneer/internal/search"
	_ "github.com/woliveiras/bookaneer/internal/search/newznab"
	_ "github.com/woliveiras/bookaneer/internal/search/torznab"
	"github.com/woliveiras/bookaneer/internal/wanted"

	"github.com/woliveiras/bookaneer/api/v1/ws"
	"github.com/woliveiras/bookaneer/internal/notification"
	"github.com/woliveiras/bookaneer/internal/notification/webhook"
	"github.com/woliveiras/bookaneer/internal/opds"
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

	// Handle "healthcheck" subcommand for Docker HEALTHCHECK
	if flag.NArg() > 0 && flag.Arg(0) == "healthcheck" {
		return runHealthcheck()
	}

	cfg, err := config.Load(*dataDir, *configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	setupLogging(cfg.LogLevel)
	slog.Info("starting bookaneer", "version", version, "dataDir", cfg.DataDir)

	db, err := setupDatabase(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	authSvc, err := setupAuth(db, cfg)
	if err != nil {
		return err
	}

	e := setupEcho(authSvc)
	api := e.Group("/api/v1")

	if err := registerRoutes(e, api, db, cfg, authSvc); err != nil {
		return err
	}

	if err := serveFrontend(e); err != nil {
		return fmt.Errorf("setup frontend: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("listening", "address", addr)

	sc := echo.StartConfig{
		Address:         addr,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: 10 * time.Second,
	}
	if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server: %w", err)
	}

	slog.Info("goodbye")
	return nil
}

// setupDatabase opens the database and runs migrations.
func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := database.Open(cfg.DatabasePath())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := database.Migrate(db, bookaneer.MigrationsFS, "migrations"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return db, nil
}

// setupAuth initialises the auth service, ensures the API key and default admin exist.
func setupAuth(db *sql.DB, cfg *config.Config) (*auth.Service, error) {
	ctx := context.Background()
	authSvc := auth.New(db)

	if err := authSvc.EnsureAPIKey(ctx); err != nil {
		return nil, fmt.Errorf("ensure api key: %w", err)
	}

	envPassword := os.Getenv("BOOKANEER_ADMIN_PASSWORD")
	adminPassword, err := authSvc.EnsureDefaultAdmin(ctx, envPassword)
	if err != nil {
		return nil, fmt.Errorf("ensure default admin: %w", err)
	}
	if adminPassword != "" {
		logAdminCredentials(cfg, adminPassword, envPassword)
	}

	if apiKey, err := authSvc.GetAPIKey(ctx); err == nil && apiKey != "" {
		slog.Info("API key for external integrations (OPDS, scripts, etc)", "apiKey", apiKey)
	}

	return authSvc, nil
}

// logAdminCredentials logs the newly-generated admin credentials and optionally
// saves them to a file for Docker users who might miss the log output.
func logAdminCredentials(cfg *config.Config, password, envPassword string) {
	slog.Info("===========================================")
	slog.Info("Default admin user created")
	slog.Info("Username: admin")
	if envPassword != "" {
		slog.Info("Password: <set via BOOKANEER_ADMIN_PASSWORD>")
	} else {
		slog.Info("Password: " + password)
		credentialsFile := filepath.Join(cfg.DataDir, "admin_credentials.txt")
		content := fmt.Sprintf(
			"Bookaneer Default Admin Credentials\n\nUsername: admin\nPassword: %s\n\nDelete this file after you have saved these credentials.\n",
			password,
		)
		if err := os.WriteFile(credentialsFile, []byte(content), 0600); err != nil {
			slog.Warn("could not save credentials file", "error", err)
		} else {
			slog.Info("Credentials also saved to: " + credentialsFile)
		}
	}
	slog.Info("Please change your password after first login!")
	slog.Info("===========================================")
}

// setupEcho creates and configures the Echo instance with global middleware.
func setupEcho(authSvc *auth.Service) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(apimw.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-Api-Key"},
	}))

	return e
}

// registerRoutes wires all API handlers to the given router group.
func registerRoutes(e *echo.Echo, api *echo.Group, db *sql.DB, cfg *config.Config, authSvc *auth.Service) error {
	ctx := context.Background()

	// Public endpoints (registered on Echo, not on api group, to avoid
	// group-middleware leaking from the protected subgroup).
	systemHandler := handler.NewSystemHandler(version, buildTime, cfg, db)
	e.GET("/api/v1/system/status", systemHandler.Status)
	e.GET("/api/v1/system/health", systemHandler.Health)
	api.GET("/tag", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, []interface{}{})
	})

	authHandler := handler.NewAuthHandler(authSvc)
	api.POST("/auth/login", authHandler.Login)

	// Protected group
	protected := api.Group("")
	protected.Use(apimw.Auth(authSvc))

	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)

	settingsHandler := handler.NewSettingsHandler(authSvc, cfg)
	settingsHandler.Register(protected)

	// Core domain services + handlers
	authorSvc := author.New(database.OpenX(db))
	bookSvc := book.New(db)
	seriesSvc := series.New(db)
	rootFolderSvc := rootfolder.New(db)
	qualityProfileSvc := qualityprofile.New(db)
	libraryScanner := library.NewScanner(db)

	if err := qualityProfileSvc.EnsureDefault(ctx); err != nil {
		slog.Warn("could not ensure default quality profile", "error", err)
	}

	jobScheduler := scheduler.New(db, 3)

	handler.NewAuthorHandler(authorSvc).Register(protected)
	handler.NewBookHandler(bookSvc).Register(protected)
	handler.NewSeriesHandler(seriesSvc).Register(protected)
	handler.NewRootFolderHandler(rootFolderSvc, libraryScanner).Register(protected)
	handler.NewQualityProfileHandler(qualityProfileSvc).Register(protected)
	handler.NewLibraryHandler(libraryScanner).Register(protected)
	handler.NewReaderHandler(reader.New(db)).Register(protected)

	// Remote path mappings
	pathMappingSvc := pathmapping.New(db)
	handler.NewRemotePathMappingHandler(pathMappingSvc).Register(protected)

	// Metadata providers
	httpClient := &http.Client{Timeout: 30 * time.Second}
	metaAggregator := metadata.NewAggregator(
		slog.Default(),
		openlibrary.New(httpClient, "Bookaneer/1.0 (https://github.com/woliveiras/bookaneer)"),
		googlebooks.New(httpClient, ""),
	)
	metadataHandler := handler.NewMetadataHandler(metaAggregator)
	protected.GET("/metadata/authors", metadataHandler.SearchAuthors)
	protected.GET("/metadata/books", metadataHandler.SearchBooks)
	protected.GET("/metadata/authors/:foreignId", metadataHandler.GetAuthor)
	protected.GET("/metadata/books/:foreignId", metadataHandler.GetBook)
	protected.GET("/metadata/isbn/:isbn", metadataHandler.LookupISBN)
	protected.GET("/metadata/providers", metadataHandler.ListProviders)

	// Search service (Newznab/Torznab indexers)
	searchSvc := search.NewService(db)
	if err := searchSvc.LoadIndexers(ctx); err != nil {
		slog.Warn("could not load indexers", "error", err)
	}
	handler.NewSearchHandler(searchSvc).Register(protected)

	// Digital library providers
	var annasProvider digitallibrary.Provider
	if cfg.FlareSolverrURL != "" {
		slog.Info("using FlareSolverr for Anna's Archive", "url", cfg.FlareSolverrURL)
		annasProvider = annas.NewWithFlareSolverr(cfg.FlareSolverrURL)
	} else {
		annasProvider = annas.New()
	}

	providers := []digitallibrary.Provider{
		gutendex.New(),
		wikisource.New(),
		aozora.New(),
		openlibrarypublic.New(),
		archive.New(),
		dominiopublico.New(),
		sitesearch.New("gutenberg-au", "gutenberg.net.au", "html"),
		sitesearch.New("gutenberg-ca", "gutenberg.ca", "html"),
		sitesearch.New("dominio-publico-fallback", "dominiopublico.gov.br", "pdf"),
		sitesearch.New("biblioteca-digital-hispanica", "bdh.bne.es", "pdf"),
		sitesearch.New("gallica", "gallica.bnf.fr", "pdf"),
		sitesearch.New("projekt-gutenberg-de", "projekt-gutenberg.org", "html"),
		sitesearch.New("baen-free-library", "baen.com", "html"),
		sitesearch.New("ccel", "ccel.org", "html"),
		sitesearch.New("sefaria", "sefaria.org", "html"),
		sitesearch.New("ctext", "ctext.org", "html"),
		sitesearch.New("sacred-texts", "sacred-texts.com", "html"),
		sitesearch.New("digital-comic-museum", "digitalcomicmuseum.com", "pdf"),
		sitesearch.New("hathitrust", "hathitrust.org", "pdf"),
		annasProvider,
		libgen.New(),
	}

	if cfg.CustomProvidersEnable {
		for _, cp := range cfg.CustomProviders {
			if cp.Name == "" || cp.Domain == "" {
				continue
			}
			providers = append(providers, sitesearch.New(cp.Name, cp.Domain, cp.FormatHint))
		}
	}

	libAggregator := digitallibrary.NewAggregator(providers...)
	handler.NewDigitalLibraryHandler(libAggregator).Register(protected)

	// Download service
	downloadSvc := download.NewService(db)
	handler.NewDownloadHandler(downloadSvc).Register(protected)

	// Wanted service
	namingEngine := naming.New(db)
	handler.NewNamingHandler(namingEngine).Register(protected)
	wantedSvc := wanted.New(db, bookSvc, libAggregator, searchSvc, downloadSvc, namingEngine, libraryScanner, pathMappingSvc)
	jobScheduler.RegisterWantedHandlers(wantedSvc)
	jobScheduler.Start(ctx)

	handler.NewWantedHandler(wantedSvc, jobScheduler).Register(protected)

	// Notification service
	notifSvc := notification.New(db)
	notifSvc.RegisterFactory("webhook", func(cfg notification.Config) (notification.Channel, error) {
		return webhook.New(cfg)
	})
	handler.NewNotificationHandler(notifSvc).Register(protected)

	// WebSocket hub
	wsHub := ws.NewHub()
	api.GET("/ws", wsHub.HandleWS)

	// Backup / Restore
	protected.GET("/system/backup", systemHandler.Backup)
	protected.POST("/system/restore", systemHandler.Restore)

	// OPDS catalog (public, uses API key auth via query param)
	opdsServer := opds.New(db)
	opdsServer.Register(e)

	// API documentation
	handler.NewDocsHandler().Register(api)

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
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})))

	return nil
}

// runHealthcheck performs a lightweight HTTP health check against the running
// instance. Used by Docker HEALTHCHECK to determine container health without
// shell or curl (scratch image).
func runHealthcheck() error {
	port := os.Getenv("BOOKANEER_PORT")
	if port == "" {
		port = "9090"
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/api/v1/system/health", port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
