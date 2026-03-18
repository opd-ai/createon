// Package templates provides HTML template rendering and markdown processing.
//
// The templates package handles rendering of HTML pages using Go's html/template
// package and converts markdown content to HTML using the goldmark library.
//
// # Features
//
//   - Template caching for improved performance
//   - GitHub Flavored Markdown (GFM) support
//   - Custom template functions for formatting
//   - Thread-safe template access
//
// # Template Functions
//
// The following template functions are available:
//   - markdownToHTML: Converts markdown string to HTML
//   - formatPrice: Formats cryptocurrency amounts with currency suffix
//   - timeAgo: Formats time as relative duration
//
// # Usage
//
// Create a Manager with template configuration:
//
//	tm, err := templates.NewManager(templates.TemplateConfig{
//	    TemplatesDir: "./templates",
//	    AssetsDir:    "./assets",
//	    BaseLayout:   "base.html",
//	    DevMode:      false,
//	})
//
// Render a page:
//
//	err = tm.RenderPage(w, "home.html", templates.PageData{
//	    Title: "Welcome",
//	})
package templates
