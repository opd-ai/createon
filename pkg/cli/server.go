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
	"github.com/opd-ai/createon/pkg/auth"
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
	auth      *auth.Manager
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

	// Parse payment timeout from config, default to 24 hours
	paymentTimeout := 24 * time.Hour
	if cfg.Paywall.Timeout != "" {
		if parsed, err := time.ParseDuration(cfg.Paywall.Timeout); err == nil {
			paymentTimeout = parsed
		}
	}

	// Initialize paywall
	pw, err := paywall.NewPaywall(paywall.Config{
		TestNet:          cfg.Paywall.TestNet,
		PriceInBTC:       cfg.Paywall.DefaultBTC,
		PriceInXMR:       cfg.Paywall.DefaultXMR,
		Store:            paywall.NewFileStore(),
		PaymentTimeout:   paymentTimeout,
		MinConfirmations: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize paywall: %w", err)
	}

	sm := subscription.NewManager(fm, pw, cfg.DataDir, cfg)
	am := auth.NewManager(fm, 24*time.Hour)

	s := &server{
		cfg:       cfg,
		files:     fm,
		templates: tm,
		subs:      sm,
		auth:      am,
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
	s.router.Use(s.auth.Middleware) // Add auth middleware

	// Static files
	fileServer := http.FileServer(http.Dir(s.cfg.AssetsDir))
	s.router.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))

	// Auth routes
	s.router.Get("/login", s.handleLoginPage())
	s.router.Post("/login", s.handleLogin())
	s.router.Get("/register", s.handleRegisterPage())
	s.router.Post("/register", s.handleRegister())
	s.router.Get("/logout", s.handleLogout())
	s.router.Get("/dashboard", s.handleDashboard())

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

		// Load creator info
		var creator createon.Creator
		if err := s.files.ReadYAML(
			filepath.Join("creators", username, "config.yaml"),
			&creator,
		); err != nil {
			http.Error(w, "Creator not found", http.StatusNotFound)
			return
		}

		// Get user from session
		user := auth.GetUserFromContext(r.Context())
		userEmail := ""
		if user != nil {
			userEmail = user.Email
		}

		// Check access
		if userEmail == "" || !s.subs.VerifyAccess(r.Context(), userEmail, username, meta.TierID) {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		html, err := s.templates.RenderMarkdown(content)
		if err != nil {
			http.Error(w, "Failed to render post", http.StatusInternalServerError)
			return
		}

		s.templates.RenderPage(w, "post.html", templates.PageData{
			Title:   meta.Title,
			Creator: &creator,
			Post:    &meta,
			Content: template.HTML(html),
		})
	}
}

// PaymentPageData holds both tier and payment info for the payment template
type PaymentPageData struct {
	Tier    *createon.Tier
	Payment interface{}
}

func (s *server) handleSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		tierID := chi.URLParam(r, "tierID")

		// Get user from session
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Load creator to get tier info
		var creator createon.Creator
		if err := s.files.ReadYAML(
			filepath.Join("creators", username, "config.yaml"),
			&creator,
		); err != nil {
			http.Error(w, "Creator not found", http.StatusNotFound)
			return
		}

		// Find the tier
		var selectedTier *createon.Tier
		for _, tier := range creator.Tiers {
			if tier.ID == tierID {
				selectedTier = &tier
				break
			}
		}
		if selectedTier == nil {
			http.Error(w, "Tier not found", http.StatusNotFound)
			return
		}

		// Create subscription
		payment, err := s.subs.CreateSubscription(r.Context(), user.Email, username, tierID)
		if err != nil {
			http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
			return
		}

		s.templates.RenderPage(w, "payment.html", templates.PageData{
			Title:   "Subscribe",
			Creator: &creator,
			Content: PaymentPageData{
				Tier:    selectedTier,
				Payment: payment,
			},
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

func (s *server) handleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.templates.RenderPage(w, "login.html", templates.PageData{
			Title: "Login",
		})
	}
}

func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		session, err := s.auth.Login(email, password)
		if err != nil {
			s.templates.RenderPage(w, "login.html", templates.PageData{
				Title: "Login",
				Flash: "Invalid email or password",
			})
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session.Token,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Redirect to dashboard or original page
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/dashboard"
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func (s *server) handleRegisterPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.templates.RenderPage(w, "register.html", templates.PageData{
			Title: "Register",
		})
	}
}

func (s *server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if err := s.auth.Register(email, password); err != nil {
			s.templates.RenderPage(w, "register.html", templates.PageData{
				Title: "Register",
				Flash: "Registration failed: " + err.Error(),
			})
			return
		}

		// Auto-login after registration
		session, err := s.auth.Login(email, password)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session.Token,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (s *server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			s.auth.Logout(cookie.Value)
		}

		// Clear session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		s.templates.RenderPage(w, "dashboard.html", templates.PageData{
			Title: "Dashboard",
			User:  &createon.User{Email: user.Email, PasswordHash: user.PasswordHash, CreatedAt: user.CreatedAt},
		})
	}
}
