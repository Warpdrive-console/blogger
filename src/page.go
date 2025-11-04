package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func postPage(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	postsMu.RLock()
	p, exists := posts[id]
	postsMu.RUnlock()
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		scheme := "http"
		if c.Request.TLS != nil || strings.EqualFold(c.Request.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		host := c.Request.Host
		baseURL = fmt.Sprintf("%s://%s", scheme, host)
	}

	pageURL := fmt.Sprintf("%s/post/%d", baseURL, p.ID)
	image := ""
	if p.Thumbnail != "" {
		if strings.HasPrefix(p.Thumbnail, "http://") || strings.HasPrefix(p.Thumbnail, "https://") {
			image = p.Thumbnail
		} else {
			image = strings.TrimRight(baseURL, "/") + p.Thumbnail
		}
	}

	meta := fmt.Sprintf(`
<meta property="og:type" content="article" />
<meta property="og:url" content="%s" />
<meta property="og:title" content="%s" />
<meta property="og:description" content="%s" />
<meta name="twitter:card" content="summary_large_image" />
<meta name="twitter:title" content="%s" />
<meta name="twitter:description" content="%s" />
<meta name="author" content="%s" />
<link rel="canonical" href="%s" />
`, pageURL, p.Title, p.Description, p.Title, p.Description, p.Author, pageURL)

	if image != "" {
		meta += fmt.Sprintf(`<meta property="og:image" content="%s" />
<meta name="twitter:image" content="%s" />`, image, image)
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../static"
	}

	tmpl, err := template.ParseFiles(filepath.Join(staticDir, "post.html"))
	if err != nil {
		c.String(http.StatusInternalServerError, "template error: %v", err)
		return
	}

	data := map[string]any{
		"Title":    p.Title,
		"MetaTags": template.HTML(meta),
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(c.Writer, data)
}
