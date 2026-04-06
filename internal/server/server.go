package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-estimate/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/estimates", s.listEstimates)
	s.mux.HandleFunc("POST /api/estimates", s.createEstimates)
	s.mux.HandleFunc("GET /api/estimates/export.csv", s.exportEstimates)
	s.mux.HandleFunc("GET /api/estimates/{id}", s.getEstimates)
	s.mux.HandleFunc("PUT /api/estimates/{id}", s.updateEstimates)
	s.mux.HandleFunc("DELETE /api/estimates/{id}", s.delEstimates)
	s.mux.HandleFunc("GET /api/line_items", s.listLineItems)
	s.mux.HandleFunc("POST /api/line_items", s.createLineItems)
	s.mux.HandleFunc("GET /api/line_items/export.csv", s.exportLineItems)
	s.mux.HandleFunc("GET /api/line_items/{id}", s.getLineItems)
	s.mux.HandleFunc("PUT /api/line_items/{id}", s.updateLineItems)
	s.mux.HandleFunc("DELETE /api/line_items/{id}", s.delLineItems)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listEstimates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"estimates": oe(s.db.SearchEstimates(q, filters))}); return }
	wj(w, 200, map[string]any{"estimates": oe(s.db.ListEstimates())})
}

func (s *Server) createEstimates(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Estimates
	json.NewDecoder(r.Body).Decode(&e)
	if e.ClientName == "" { we(w, 400, "client_name required"); return }
	if e.Title == "" { we(w, 400, "title required"); return }
	s.db.CreateEstimates(&e)
	wj(w, 201, s.db.GetEstimates(e.ID))
}

func (s *Server) getEstimates(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetEstimates(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateEstimates(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetEstimates(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Estimates
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.ClientName == "" { patch.ClientName = existing.ClientName }
	if patch.ClientEmail == "" { patch.ClientEmail = existing.ClientEmail }
	if patch.ClientPhone == "" { patch.ClientPhone = existing.ClientPhone }
	if patch.Title == "" { patch.Title = existing.Title }
	if patch.Description == "" { patch.Description = existing.Description }
	if patch.ValidUntil == "" { patch.ValidUntil = existing.ValidUntil }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateEstimates(&patch)
	wj(w, 200, s.db.GetEstimates(patch.ID))
}

func (s *Server) delEstimates(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteEstimates(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportEstimates(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListEstimates()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=estimates.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "client_name", "client_email", "client_phone", "title", "description", "total", "valid_until", "status", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.ClientName), fmt.Sprintf("%v", e.ClientEmail), fmt.Sprintf("%v", e.ClientPhone), fmt.Sprintf("%v", e.Title), fmt.Sprintf("%v", e.Description), fmt.Sprintf("%v", e.Total), fmt.Sprintf("%v", e.ValidUntil), fmt.Sprintf("%v", e.Status), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listLineItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"line_items": oe(s.db.SearchLineItems(q, filters))}); return }
	wj(w, 200, map[string]any{"line_items": oe(s.db.ListLineItems())})
}

func (s *Server) createLineItems(w http.ResponseWriter, r *http.Request) {
	var e store.LineItems
	json.NewDecoder(r.Body).Decode(&e)
	if e.EstimateId == "" { we(w, 400, "estimate_id required"); return }
	if e.Description == "" { we(w, 400, "description required"); return }
	s.db.CreateLineItems(&e)
	wj(w, 201, s.db.GetLineItems(e.ID))
}

func (s *Server) getLineItems(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetLineItems(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateLineItems(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetLineItems(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.LineItems
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.EstimateId == "" { patch.EstimateId = existing.EstimateId }
	if patch.Description == "" { patch.Description = existing.Description }
	s.db.UpdateLineItems(&patch)
	wj(w, 200, s.db.GetLineItems(patch.ID))
}

func (s *Server) delLineItems(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteLineItems(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportLineItems(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListLineItems()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=line_items.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "estimate_id", "description", "quantity", "unit_price", "total", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.EstimateId), fmt.Sprintf("%v", e.Description), fmt.Sprintf("%v", e.Quantity), fmt.Sprintf("%v", e.UnitPrice), fmt.Sprintf("%v", e.Total), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["estimates_total"] = s.db.CountEstimates()
	m["line_items_total"] = s.db.CountLineItems()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "estimate"}
	m["estimates"] = s.db.CountEstimates()
	m["line_items"] = s.db.CountLineItems()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}
