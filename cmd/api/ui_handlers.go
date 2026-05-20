package main

import (
	"net/http"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/query"
	"github.com/bisheshops/dynamic-crm-engine/internal/response"
	"github.com/go-chi/chi/v5"
)

type PageData struct {
	Entities []database.BatchEntity
}

func (api *API) DashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	req := query.Request{
		Limit:  50,
		Offset: 0,
	}

	results, err := api.DB.QueryEntities(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to load dashboard data", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Entities: results,
	}

	ts, ok := api.Cache["dashboard.html"]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	response.HTML(w, http.StatusOK, ts, "base", data)
}

func (api *API) EntityTableFragmentHandler(w http.ResponseWriter, r *http.Request) {
	req := query.Request{
		Limit:  50,
		Offset: 0,
	}

	results, err := api.DB.QueryEntities(r.Context(), req)
	if err != nil {
		http.Error(w, "<tr><td colspan='4'>Failed to refresh data</td></tr>", http.StatusInternalServerError)
		return
	}

	ts, ok := api.Cache["dashboard.html"]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	response.HTML(w, http.StatusOK, ts, "entity_rows", results)
}

func (api *API) DeleteEntityUIHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := api.DB.DeleteEntityByID(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
