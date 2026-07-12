package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"auction/auction/internal/middleware"
	"auction/auction/internal/model"
	"auction/auction/internal/repository"
	"auction/auction/internal/service"
)

type Handler struct{ service *service.Service }

func New(s *service.Service, auth *middleware.Auth) http.Handler {
	h := &Handler{service: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/products", h.list)
	mux.Handle("POST /api/products", auth.Protect(http.HandlerFunc(h.create)))
	mux.HandleFunc("GET /api/products/{product_id}", h.get)
	mux.Handle("PUT /api/products/{product_id}", auth.Protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/products/{product_id}", auth.Protect(http.HandlerFunc(h.cancel)))
	mux.Handle("GET /api/me/products", auth.Protect(http.HandlerFunc(h.mine)))
	mux.Handle("POST /api/products/{product_id}/bids", auth.Protect(http.HandlerFunc(h.bid)))
	mux.Handle("GET /api/products/{product_id}/bids", auth.Protect(http.HandlerFunc(h.bids)))
	mux.Handle("POST /api/products/{product_id}/sell", auth.Protect(http.HandlerFunc(h.sell)))
	mux.Handle("POST /api/uploads/presigned-url", auth.Protect(http.HandlerFunc(h.upload)))
	return mux
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in service.ProductInput
	if decode(w, r, &in) != nil {
		bad(w)
		return
	}
	p, err := h.service.Create(r.Context(), userID(r), in)
	respond(w, http.StatusCreated, p, err)
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.List(r.Context())
	respond(w, 200, v, err)
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Get(r.Context(), r.PathValue("product_id"))
	respond(w, 200, v, err)
}
func (h *Handler) mine(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Mine(r.Context(), userID(r))
	respond(w, 200, v, err)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var in service.ProductInput
	if decode(w, r, &in) != nil {
		bad(w)
		return
	}
	v, err := h.service.Update(r.Context(), r.PathValue("product_id"), userID(r), in)
	respond(w, 200, v, err)
}
func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	err := h.service.Cancel(r.Context(), r.PathValue("product_id"), userID(r))
	if err == nil {
		w.WriteHeader(204)
		return
	}
	respond(w, 200, nil, err)
}

type bidRequest struct {
	Amount model.Money `json:"amount"`
}

func (h *Handler) bid(w http.ResponseWriter, r *http.Request) {
	var in bidRequest
	if decode(w, r, &in) != nil {
		bad(w)
		return
	}
	v, err := h.service.Bid(r.Context(), r.PathValue("product_id"), userID(r), in.Amount)
	respond(w, 201, v, err)
}
func (h *Handler) bids(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Bids(r.Context(), r.PathValue("product_id"), userID(r))
	respond(w, 200, v, err)
}
func (h *Handler) sell(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.Sell(r.Context(), r.PathValue("product_id"), userID(r))
	respond(w, 200, v, err)
}

type uploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	var in uploadRequest
	if decode(w, r, &in) != nil {
		bad(w)
		return
	}
	v, err := h.service.Upload(r.Context(), userID(r), in.Filename, in.ContentType)
	respond(w, 200, v, err)
}
func userID(r *http.Request) string {
	claims, _ := middleware.Claims(r.Context())
	return claims.Subject
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values")
	}
	return nil
}
func bad(w http.ResponseWriter) { writeError(w, 400, "invalid_request", "request body is invalid") }
func respond(w http.ResponseWriter, status int, value any, err error) {
	if err == nil {
		writeJSON(w, status, value)
		return
	}
	switch {
	case errors.Is(err, service.ErrInvalid):
		writeError(w, 400, "invalid_input", "input is invalid")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, 404, "not_found", "product not found")
	case errors.Is(err, repository.ErrForbidden):
		writeError(w, 403, "forbidden", "operation is not allowed")
	case errors.Is(err, repository.ErrConflict):
		writeError(w, 409, "conflict", "auction state does not allow this operation")
	default:
		writeError(w, 500, "internal_error", "internal server error")
	}
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
