package frontend

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// APPLICATION CONFIGURATION
// ============================================================

const (
	DefaultAppName        = "ChatApplication"
	DefaultDescription    = "Secure Chat Application"
	DefaultPageTitle      = "ChatApplication"
	DefaultNotFoundTitle  = "Page Not Found"
	DefaultNotFoundStatus = http.StatusNotFound
)

// ============================================================
// USER VIEW
// ============================================================

type UserView struct {
	ID    uint
	Name  string
	Email string
	Role  string
}

// ============================================================
// PAGE DATA
// ============================================================

type PageData struct {
	AppName string

	Title       string
	Description string

	Year int

	User *UserView

	IsAuthenticated bool

	Data any
}

// ============================================================
// RENDERER
// ============================================================

type Renderer struct {
	publicDir string

	appName     string
	description string

	// devMode: when true, templates are re-parsed on every request.
	devMode bool

	mu    sync.RWMutex
	cache map[string]*template.Template
}

// ============================================================
// NEW RENDERER
// ============================================================

func NewRenderer(publicDir string) (*Renderer, error) {

	absolutePath, err := filepath.Abs(publicDir)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve frontend directory: %w",
			err,
		)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return nil, fmt.Errorf(
			"frontend directory %s: %w",
			absolutePath,
			err,
		)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"frontend path is not a directory: %s",
			absolutePath,
		)
	}

	return &Renderer{
		publicDir: absolutePath,

		appName:     DefaultAppName,
		description: DefaultDescription,

		cache: make(map[string]*template.Template),
	}, nil
}

// ============================================================
// PUBLIC DIRECTORY
// ============================================================

func (r *Renderer) PublicDir() string {

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.publicDir
}

// ============================================================
// APPLICATION INFORMATION
// ============================================================

func (r *Renderer) SetAppName(name string) {

	name = strings.TrimSpace(name)

	if name == "" {
		name = DefaultAppName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.appName = name
}

func (r *Renderer) SetDescription(description string) {

	description = strings.TrimSpace(description)

	if description == "" {
		description = DefaultDescription
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.description = description
}

func (r *Renderer) applicationInfo() (string, string) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.appName, r.description
}

// ============================================================
// DEV MODE
// ============================================================

// SetDevMode disables template caching so HTML edits show up on refresh.
func (r *Renderer) SetDevMode(enabled bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.devMode = enabled
}

func (r *Renderer) isDevMode() bool {

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.devMode
}

// ============================================================
// STARTUP CHECK
// ============================================================

// CheckPages logs whether each expected page template exists.
// Call it once at startup.
func (r *Renderer) CheckPages(pages ...string) {

	for _, page := range pages {

		p := filepath.Join(
			r.publicDir,
			"templates",
			"pages",
			page+".html",
		)

		if _, err := os.Stat(p); err != nil {
			log.Printf("WARNING: page template missing: %s", p)
		} else {
			log.Printf("page template OK: %s", p)
		}
	}
}

// ============================================================
// REGISTER FRONTEND ROUTES
// ============================================================
//
// IMPORTANT:
//
// Do NOT use:
//     router.GET("/*path", ...)
//
// Gin will panic because root-level wildcard routes conflict
// with /api, /css, /js, /assets, etc.
//
// Instead, use NoRoute() as the frontend fallback.
//
// ============================================================

func (r *Renderer) Register(router *gin.Engine) {

	// --------------------------------------------------------
	// STATIC FILES
	// --------------------------------------------------------

	publicDir := r.PublicDir()

	cssDir := filepath.Join(
		publicDir,
		"css",
	)

	jsDir := filepath.Join(
		publicDir,
		"js",
	)

	assetsDir := filepath.Join(
		publicDir,
		"assets",
	)

	// CSS
	router.Static(
		"/css",
		cssDir,
	)

	// JavaScript
	router.Static(
		"/js",
		jsDir,
	)

	// Images/assets
	router.Static(
		"/assets",
		assetsDir,
	)

	// Favicon
	faviconPath := filepath.Join(
		publicDir,
		"favicon.ico",
	)

	if _, err := os.Stat(faviconPath); err == nil {

		router.StaticFile(
			"/favicon.ico",
			faviconPath,
		)
	}

	// --------------------------------------------------------
	// HOME
	// --------------------------------------------------------

	router.GET(
		"/",
		func(c *gin.Context) {

			r.Render(
				c,
				"home",
				PageData{
					Title: "Home",
				},
			)
		},
	)

	// --------------------------------------------------------
	// FRONTEND FALLBACK
	// --------------------------------------------------------

	router.NoRoute(
		r.PageHandler(),
	)
}

// ============================================================
// RENDER
// ============================================================

func (r *Renderer) Render(
	c *gin.Context,
	page string,
	data PageData,
) {

	tmpl, err := r.templateForPage(page)

	if err != nil {

		log.Printf(
			"frontend render error for page %q: %v",
			page,
			err,
		)

		c.Error(err)

		c.String(
			http.StatusInternalServerError,
			"template rendering error: %v",
			err,
		)

		return
	}

	appName, description := r.applicationInfo()

	data.AppName = appName

	data.Year = time.Now().Year()

	if data.Description == "" {
		data.Description = description
	}

	if data.Title == "" {
		data.Title = pageTitle(page)
	}

	// FIX: requests that reach NoRoute() have their status preset
	// to 404 by Gin. Set 200 explicitly so successfully rendered
	// pages (like /register) are not reported as 404.
	c.Status(http.StatusOK)

	err = tmpl.ExecuteTemplate(
		c.Writer,
		"base",
		data,
	)

	if err != nil {

		log.Printf(
			"frontend execute error for page %q: %v",
			page,
			err,
		)

		c.Error(err)

		c.String(
			http.StatusInternalServerError,
			"template rendering error",
		)

		return
	}
}

// ============================================================
// TEMPLATE LOADING
// ============================================================

func (r *Renderer) templateForPage(
	page string,
) (*template.Template, error) {

	page = strings.TrimSpace(page)

	page = strings.TrimPrefix(
		page,
		"/",
	)

	page = strings.TrimSuffix(
		page,
		"/",
	)

	if page == "" {
		page = "home"
	}

	// --------------------------------------------------------
	// SECURITY
	// --------------------------------------------------------

	if strings.Contains(page, "..") {
		return nil, fmt.Errorf(
			"invalid frontend page path",
		)
	}

	if strings.Contains(page, "\\") {
		return nil, fmt.Errorf(
			"invalid frontend page path",
		)
	}

	if !isSafePageName(page) {
		return nil, fmt.Errorf(
			"invalid frontend page name: %s",
			page,
		)
	}

	// --------------------------------------------------------
	// CACHE (skipped in dev mode)
	// --------------------------------------------------------

	devMode := r.isDevMode()

	if !devMode {

		r.mu.RLock()

		cached, exists := r.cache[page]

		r.mu.RUnlock()

		if exists {
			return cached, nil
		}
	}

	// --------------------------------------------------------
	// TEMPLATE PATHS
	// --------------------------------------------------------

	basePath := filepath.Join(
		r.publicDir,
		"templates",
		"layouts",
		"base.html",
	)

	navbarPath := filepath.Join(
		r.publicDir,
		"templates",
		"components",
		"navbar.html",
	)

	footerPath := filepath.Join(
		r.publicDir,
		"templates",
		"components",
		"footer.html",
	)

	pagePath := filepath.Join(
		r.publicDir,
		"templates",
		"pages",
		page+".html",
	)

	// --------------------------------------------------------
	// VERIFY FILES
	// --------------------------------------------------------

	files := []string{
		basePath,
		navbarPath,
		footerPath,
		pagePath,
	}

	for _, path := range files {

		info, err := os.Stat(path)

		if err != nil {
			return nil, fmt.Errorf(
				"template file not found %s: %w",
				path,
				err,
			)
		}

		if info.IsDir() {
			return nil, fmt.Errorf(
				"template path is a directory: %s",
				path,
			)
		}
	}

	// --------------------------------------------------------
	// PARSE
	// --------------------------------------------------------

	tmpl, err := template.
		New("base").
		Option("missingkey=error").
		ParseFiles(
			basePath,
			navbarPath,
			footerPath,
			pagePath,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"parse frontend templates: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// CACHE (skipped in dev mode)
	// --------------------------------------------------------

	if !devMode {

		r.mu.Lock()

		r.cache[page] = tmpl

		r.mu.Unlock()
	}

	return tmpl, nil
}

// ============================================================
// PAGE NAME SECURITY
// ============================================================

func isSafePageName(page string) bool {

	if page == "" {
		return false
	}

	for _, char := range page {

		switch {

		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '-':
		case char == '_':
		case char == '/':
		default:
			return false
		}
	}

	return true
}

// ============================================================
// FRONTEND PAGE HANDLER
// ============================================================

func (r *Renderer) PageHandler() gin.HandlerFunc {

	return func(c *gin.Context) {

		path := c.Request.URL.Path

		// ====================================================
		// API MUST NEVER FALL INTO FRONTEND
		// ====================================================

		if path == "/api" ||
			strings.HasPrefix(path, "/api/") {

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "API endpoint not found",
				},
			)

			return
		}

		// ====================================================
		// STATIC FILE PROTECTION
		// ====================================================

		if path == "/css" ||
			strings.HasPrefix(path, "/css/") ||
			path == "/js" ||
			strings.HasPrefix(path, "/js/") ||
			path == "/assets" ||
			strings.HasPrefix(path, "/assets/") {

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "static file not found",
				},
			)

			return
		}

		// ====================================================
		// HOME
		// ====================================================

		if path == "/" || path == "" {

			r.Render(
				c,
				"home",
				PageData{
					Title: "Home",
				},
			)

			return
		}

		// ====================================================
		// NORMALIZE
		// ====================================================

		path = strings.Trim(
			path,
			"/",
		)

		if path == "" {

			r.Render(
				c,
				"home",
				PageData{
					Title: "Home",
				},
			)

			return
		}

		// ====================================================
		// SECURITY
		// ====================================================

		if !isSafePageName(path) {

			log.Printf(
				"frontend 404: unsafe page name %q",
				c.Request.URL.Path,
			)

			r.notFound(c)

			return
		}

		// ====================================================
		// ONLY ALLOW PAGE PATHS
		// ====================================================

		pagePath := filepath.Join(
			r.publicDir,
			"templates",
			"pages",
			path+".html",
		)

		info, err := os.Stat(pagePath)

		if err != nil {

			if os.IsNotExist(err) {

				log.Printf(
					"frontend 404: no template for %q (looked for %s)",
					c.Request.URL.Path,
					pagePath,
				)

				r.notFound(c)

				return
			}

			c.Error(err)

			c.String(
				http.StatusInternalServerError,
				"frontend error",
			)

			return
		}

		if info.IsDir() {

			r.notFound(c)

			return
		}

		// ====================================================
		// RENDER PAGE
		// ====================================================

		r.Render(
			c,
			path,
			PageData{
				Title: pageTitle(path),
			},
		)
	}
}

// ============================================================
// 404
// ============================================================

func (r *Renderer) notFound(
	c *gin.Context,
) {

	// Avoid recursively trying to render 404 if 404 itself
	// does not exist.
	tmpl, err := r.templateForPage("404")

	if err != nil {

		c.String(
			DefaultNotFoundStatus,
			"404 - Page not found",
		)

		return
	}

	appName, description := r.applicationInfo()

	data := PageData{
		AppName: appName,

		Title: DefaultNotFoundTitle,

		Description: description,

		Year: time.Now().Year(),
	}

	c.Status(
		DefaultNotFoundStatus,
	)

	err = tmpl.ExecuteTemplate(
		c.Writer,
		"base",
		data,
	)

	if err != nil {

		c.String(
			DefaultNotFoundStatus,
			"404 - Page not found",
		)
	}
}

// ============================================================
// PAGE TITLE
// ============================================================

func pageTitle(page string) string {

	page = strings.TrimSpace(page)

	if page == "" || page == "home" {
		return DefaultPageTitle
	}

	page = strings.ReplaceAll(
		page,
		"-",
		" ",
	)

	page = strings.ReplaceAll(
		page,
		"_",
		" ",
	)

	page = strings.ReplaceAll(
		page,
		"/",
		" ",
	)

	words := strings.Fields(page)

	for i, word := range words {

		if word == "" {
			continue
		}

		runes := []rune(word)

		if len(runes) == 0 {
			continue
		}

		runes[0] = []rune(
			strings.ToUpper(
				string(runes[0]),
			),
		)[0]

		words[i] = string(runes)
	}

	title := strings.Join(
		words,
		" ",
	)

	if title == "" {
		return DefaultPageTitle
	}

	return title
}
