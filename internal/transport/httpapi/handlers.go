package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/rooseveltrp/go-url-shortener/internal/shortener"
	"github.com/rooseveltrp/go-url-shortener/internal/storage"
)

type Server struct {
	store   *storage.Store
	baseURL string
}

func NewServer(store *storage.Store, baseURL string) *Server {
	return &Server{store: store, baseURL: strings.TrimRight(baseURL, "/")}
}

type shortenReq struct {
	URL    string `json:"url"`
	Custom string `json:"custom,omitempty"`
}

type shortenResp struct {
	Code    string `json:"code"`
	Short   string `json:"short_url"`
	Original string `json:"original_url"`
}

func (s *Server) HandleShorten(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := validateURL(req.URL); err != nil {
		httpError(w, http.StatusBadRequest, "invalid URL: "+err.Error())
		return
	}

	code := strings.TrimSpace(req.Custom)
	if code != "" {
		// ensure unique
		exists, err := s.store.Exists(code)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "store error")
			return
		}
		if exists {
			httpError(w, http.StatusConflict, "custom code already exists")
			return
		}
	} else {
		// generate unique code
		for {
			c, err := shortener.GenerateCode(6)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "code generation failed")
				return
			}
			exists, err := s.store.Exists(c)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "store error")
				return
			}
			if !exists {
				code = c
				break
			}
		}
	}

	if err := s.store.Save(code, req.URL); err != nil {
		httpError(w, http.StatusInternalServerError, "failed to save")
		return
	}

	resp := shortenResp{
		Code:     code,
		Short:    s.baseURL + "/" + code,
		Original: req.URL,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	u, err := s.store.Get(code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		httpError(w, http.StatusInternalServerError, "store error")
		return
	}
	_ = s.store.IncHit(code)
	http.Redirect(w, r, u, http.StatusFound)
}

type getURLResp struct {
	Code    string `json:"code"`
	URL     string `json:"url"`
	Hits    uint64 `json:"hits"`
	Short   string `json:"short_url"`
}

func (s *Server) HandleGetURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		httpError(w, http.StatusBadRequest, "missing code")
		return
	}
	u, err := s.store.Get(code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpError(w, http.StatusNotFound, "not found")
			return
		}
		httpError(w, http.StatusInternalServerError, "store error")
		return
	}
	hits, _ := s.store.Hits(code)
	writeJSON(w, http.StatusOK, getURLResp{
		Code:  code,
		URL:   u,
		Hits:  hits,
		Short: s.baseURL + "/" + code,
	})
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("scheme must be http or https")
	}
	if u.Host == "" {
		return errors.New("missing host")
	}
	return nil
}

func httpError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
