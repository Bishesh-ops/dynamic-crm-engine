package ui

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/bisheshops/dynamic-crm-engine/web"
)

type TemplateCache map[string]*template.Template

func NewTemplateCache() (TemplateCache, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(web.Assets, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("CRITICAL: no templates found in embedded filesystem")
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFS(
			web.Assets,
			"templates/base.html",
			page,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		ts, err = ts.ParseFS(web.Assets, "templates/partials/*.html")
		if err != nil {
			return nil, fmt.Errorf("failed to parse partials: %w", err)
		}

		cache[name] = ts
	}

	return cache, nil
}
