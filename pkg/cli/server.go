// pkg/cli/server.go
package cli

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"

	"github.com/opd-ai/createon"
	"github.com/opd-ai/createon/pkg/files"
	"github.com/opd-ai/createon/pkg/subscription"
	"github.com/opd-ai/createon/pkg/templates"
	"github.com/opd-ai/paywall"
)

func init() {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run the Createon server",
		RunE:  runServer,
	}

	// Server flags
	serverCmd.Flags().StringP("host", "H", "localhost", "Server host")
	serverCmd.Flags().IntP("port", "p", 8080, "Server port")
	serverCmd.Flags().StringP("data", "d", "data", "Data directory")
	serverCmd.Flags().Bool("dev", false, "Development mode")

	rootCmd.AddCommand(serverCmd)
}

type server struct {
	cfg       *createon.Config
	files     *files.Manager
	templates *templates.Manager
	subs      *subscription.Manager
	paywall   *paywall.Paywall
	router    *chi.Mux
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load config
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg, err := createon.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with flags
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		cfg.Server.Host = host
	}
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		cfg.Server.Port = port
	}
	if data, _ := cmd.Flags().GetString("data"); data != "" {
		cfg.DataDir = data
	}

	// Initialize managers
	fm, err := files.NewManager(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize file manager: %w", err)
	}

	tm, err := templates.NewManager(templates.TemplateConfig{
		TemplatesDir: cfg.TemplateDir,
		AssetsDir:    cfg.AssetsDir,
		BaseLayout:   "base.html",
		DevMode:      cfg.Paywall.TestNet,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Initialize paywall
	pw, err := paywall.NewPaywall(paywall.Config{
		TestNet:          cfg.Paywall.TestNet,
		PriceInBTC:       cfg.Paywall.DefaultBTC,
		PriceInXMR:       cfg.Paywall.DefaultXMR,
		Store:            paywall.NewFileStore(),
		PaymentTimeout:   24 * time.Hour,
		MinConfirmations: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize paywall: %w", err)
	}

	sm := subscription.NewManager(fm, pw, cfg.DataDir)

	s := &server{
		cfg:       cfg,
		files:     fm,
		templates: tm,
		subs:      sm,
		paywall:   pw,
		router:    chi.NewRouter(),
	}

	return s.serve()
}

func (s *server) serve() error {
	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.CleanPath)
	s.router.Use(middleware.GetHead)

	// Static files
	fileServer := http.FileServer(http.Dir(s.cfg.AssetsDir))
	s.router.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))

	// Routes
	s.router.Get("/", s.handleHome())
	s.router.Get("/c/{username}", s.handleCreatorProfile())
	s.router.Get("/c/{username}/p/{postID}", s.handleViewPost())
	s.router.Get("/subscribe/{username}/{tierID}", s.handleSubscribe())
	s.router.Post("/payment/verify/{paymentID}", s.handleVerifyPayment())

	// Create server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	// Start server
	log.Printf("Server is running on http://%s\n", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	log.Println("Server stopped")
	return nil
}

func (s *server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// List creators
		entries, err := os.ReadDir(filepath.Join(s.cfg.DataDir, "creators"))
		if err != nil {
			http.Error(w, "Failed to list creators", http.StatusInternalServerError)
			return
		}

		var creators []createon.Creator
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			var creator createon.Creator
			if err := s.files.ReadYAML(
				filepath.Join("creators", entry.Name(), "config.yaml"),
				&creator,
			); err != nil {
				continue
			}
			creators = append(creators, creator)
		}

		s.templates.RenderPage(w, "home.html", templates.PageData{
			Title:   "Createon - Creator Platform",
			Content: creators,
		})
	}
}

func (s *server) handleCreatorProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")

		// Load creator
		var creator createon.Creator
		if err := s.files.ReadYAML(
			filepath.Join("creators", username, "config.yaml"),
			&creator,
		); err != nil {
			http.Error(w, "Creator not found", http.StatusNotFound)
			return
		}

		s.templates.RenderPage(w, "profile.html", templates.PageData{
			Title:   creator.DisplayName,
			Creator: &creator,
		})
	}
}

func (s *server) handleViewPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		postID := chi.URLParam(r, "postID")

		// Load post
		content, err := s.files.ReadFile(
			filepath.Join("creators", username, "posts", postID+".md"),
		)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var meta createon.Post
		if err := s.files.ReadYAML(
			filepath.Join("creators", username, "posts", "metadata.yaml"),
			&meta,
		); err != nil {
			http.Error(w, "Post metadata not found", http.StatusNotFound)
			return
		}

		// Check access
		// TODO: Implement user session management
		if !s.subs.VerifyAccess(r.Context(), "user@example.com", username, meta.TierID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		html, err := s.templates.RenderMarkdown(content)
		if err != nil {
			http.Error(w, "Failed to render post", http.StatusInternalServerError)
			return
		}

		s.templates.RenderPage(w, "post.html", templates.PageData{
			Title:   meta.Title,
			Content: template.HTML(html),
		})
	}
}

func (s *server) handleSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		tierID := chi.URLParam(r, "tierID")

		// Create subscription
		// TODO: Get user email from session
		payment, err := s.subs.CreateSubscription(r.Context(), "user@example.com", username, tierID)
		if err != nil {
			http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
			return
		}

		s.templates.RenderPage(w, "payment.html", templates.PageData{
			Title:   "Subscribe",
			Content: payment,
		})
	}
}

func (s *server) handleVerifyPayment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paymentID := chi.URLParam(r, "paymentID")

		if err := s.subs.ProcessPayment(r.Context(), paymentID); err != nil {
			http.Error(w, "Payment verification failed", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
