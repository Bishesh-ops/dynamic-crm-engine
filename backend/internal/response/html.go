package response

import (
	"bytes"
	"html/template"
	"net/http"
)

func HTML(w http.ResponseWriter, status int, ts *template.Template, templateName string, data any) {
	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}
