// pkg/templates/manager.go
package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	. "github.com/opd-ai/createon"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

// Manager handles template rendering and caching
type Manager struct {
	templates  *template.Template
	markdown   goldmark.Markdown
	mu         sync.RWMutex
	cache      map[string]*template.Template
	baseLayout string
}

// TemplateConfig holds template configuration
type TemplateConfig struct {
	TemplatesDir string
	AssetsDir    string
	BaseLayout   string
	DevMode      bool
}

// NewManager creates a new template manager
func NewManager(cfg TemplateConfig) (*Manager, error) {
	// Initialize goldmark with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			mdhtml.WithHardWraps(),
			mdhtml.WithXHTML(),
			mdhtml.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	// Load base templates
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"markdownToHTML": func(input string) template.HTML {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				return template.HTML("Error converting markdown")
			}
			return template.HTML(buf.String())
		},
		"formatPrice": func(amount float64, currency string) string {
			return fmt.Sprintf("%.8f %s", amount, currency)
		},
		"timeAgo": func(t time.Time) string {
			return time.Since(t).Round(time.Second).String()
		},
	}).ParseGlob(filepath.Join(cfg.TemplatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &Manager{
		templates:  tmpl,
		markdown:   md,
		cache:      make(map[string]*template.Template),
		baseLayout: cfg.BaseLayout,
	}, nil
}

// PageData holds common page data for template rendering
type PageData struct {
	Title       string
	Description string
	Creator     *Creator
	Post        *Post
	User        *User
	Content     interface{}
	Flash       string
}

// RenderPage renders a template with the given data
func (m *Manager) RenderPage(w http.ResponseWriter, name string, data PageData) error {
	m.mu.RLock()
	tmpl, exists := m.cache[name]
	m.mu.RUnlock()

	if !exists {
		var err error
		tmpl, err = m.templates.Clone()
		if err != nil {
			return fmt.Errorf("failed to clone templates: %w", err)
		}

		m.mu.Lock()
		m.cache[name] = tmpl
		m.mu.Unlock()
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, m.baseLayout, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}

// RenderMarkdown converts markdown content to HTML
func (m *Manager) RenderMarkdown(content []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := m.markdown.Convert(content, &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// RenderPost renders a markdown post to HTML
func (m *Manager) RenderPost(w http.ResponseWriter, post *Post, data PageData) error {
	// Read the markdown file
	content, err := m.RenderMarkdown([]byte(post.Content))
	if err != nil {
		return err
	}

	data.Content = content
	return m.RenderPage(w, "post.html", data)
}

// RenderCreatorProfile renders a creator's profile page
func (m *Manager) RenderCreatorProfile(w http.ResponseWriter, creator *Creator, data PageData) error {
	data.Creator = creator
	return m.RenderPage(w, "profile.html", data)
}

// RenderPaymentPage renders the payment/subscription page
func (m *Manager) RenderPaymentPage(w http.ResponseWriter, creator *Creator, tier *Tier, data PageData) error {
	data.Creator = creator
	data.Content = tier
	return m.RenderPage(w, "payment.html", data)
}
