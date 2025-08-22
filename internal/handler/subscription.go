package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"subscriptions-api/internal/model"
	"subscriptions-api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Repo *storage.Repository
	Log  *slog.Logger
}

func New(r *storage.Repository, lg *slog.Logger) *Handler {
	return &Handler{Repo: r, Log: lg}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Get("/summary", h.summary)
	})
}

// POST /subscriptions
// Create subscription
// @Summary      Create subscription
// @Description  Создать новую подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        payload  body      model.SubscriptionPayload  true  "Subscription data"
// @Success      201      {object}  map[string]string			"Created"
// @Failure      400      {object}  map[string]string			"Bad request"
// @Failure      500      {object}  map[string]string			"internal error"
// @Router       /subscriptions [post]...
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var p model.SubscriptionPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p.ServiceName = strings.TrimSpace(p.ServiceName)
	if p.ServiceName == "" || p.Price < 0 || p.UserID == "" || p.StartDate == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad user_id")
		return
	}
	start, err := parseMonthYear(p.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad start_date, use MM-YYYY")
		return
	}
	var end *time.Time
	if p.EndDate != nil && *p.EndDate != "" {
		e, err := parseMonthYear(*p.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad end_date, use MM-YYYY")
			return
		}
		end = &e
		if end.Before(start) {
			writeError(w, http.StatusBadRequest, "end_date before start_date")
			return
		}
	}

	s := &model.Subscription{
		ServiceName: p.ServiceName,
		Price:       p.Price,
		UserID:      uid,
		StartDate:   start,
		EndDate:     end,
	}
	id, err := h.Repo.Create(r.Context(), s)
	if err != nil {
		h.Log.Error("create", slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// GET /subscriptions/{id}
// Get subscription by ID
// @Summary      Get subscription
// @Description  Получить подписку по идентификатору
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Not found"
// @Failure      500  {object}  map[string]string  "Internal error"
// @Router       /subscriptions/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad id")
		return
	}
	s, err := h.Repo.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if s == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// PUT /subscriptions/{id}
// Update subscription
// @Summary      Update subscription
// @Description  Обновить подписку по идентификатору
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "UUID подписки"
// @Param        payload  body      model.SubscriptionPayload  true  "Новые значения полей"
// @Success      200      {object}  model.Subscription
// @Failure      400      {object}  map[string]string  "Bad request"
// @Failure      404      {object}  map[string]string  "Not found"
// @Failure      500      {object}  map[string]string  "Internal error"
// @Router       /subscriptions/{id} [put]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad id")
		return
	}

	var p model.SubscriptionPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p.ServiceName = strings.TrimSpace(p.ServiceName)
	if p.ServiceName == "" || p.Price < 0 || p.UserID == "" || p.StartDate == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad user_id")
		return
	}
	start, err := parseMonthYear(p.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad start_date, use MM-YYYY")
		return
	}
	var end *time.Time
	if p.EndDate != nil && *p.EndDate != "" {
		e, err := parseMonthYear(*p.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad end_date, use MM-YYYY")
			return
		}
		end = &e
		if end.Before(start) {
			writeError(w, http.StatusBadRequest, "end_date before start_date")
			return
		}
	}

	s := &model.Subscription{
		ServiceName: p.ServiceName,
		Price:       p.Price,
		UserID:      uid,
		StartDate:   start,
		EndDate:     end,
	}
	ok, err := h.Repo.Update(r.Context(), id, s)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE /subscriptions/{id}
// Delete subscription
// @Summary      Delete subscription
// @Description  Удалить подписку по идентификатору
// @Tags         subscriptions
// @Param        id   path      string  true  "UUID подписки"
// @Success      204  {string}  string  "No Content"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Not found"
// @Failure      500  {object}  map[string]string  "Internal error"
// @Router       /subscriptions/{id} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad id")
		return
	}
	ok, err := h.Repo.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /subscriptions
// List subscriptions
// @Summary      List subscriptions
// @Description  Список подписок с фильтрами и пагинацией
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  false  "Фильтр по UUID пользователя"
// @Param        service_name  query     string  false  "Фильтр по названию сервиса"
// @Param        limit         query     int     false  "Количество записей (default 20, max 100)"
// @Param        offset        query     int     false  "Смещение от начала списка"
// @Success      200           {array}   model.Subscription
// @Failure      400           {object}  map[string]string  "Bad request"
// @Failure      500           {object}  map[string]string  "Internal error"
// @Router       /subscriptions [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var (
		uid     *uuid.UUID
		service *string
		limit   = 50
		offset  = 0
	)
	if s := strings.TrimSpace(q.Get("user_id")); s != "" {
		u, err := uuid.Parse(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad user_id")
			return
		}
		uid = &u
	}
	if s := strings.TrimSpace(q.Get("service_name")); s != "" {
		service = &s
	}
	if s := strings.TrimSpace(q.Get("limit")); s != "" {
		if v, err := atoi(s); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if s := strings.TrimSpace(q.Get("offset")); s != "" {
		if v, err := atoi(s); err == nil && v >= 0 {
			offset = v
		}
	}

	items, err := h.Repo.List(r.Context(), storage.ListFilter{UserID: uid, ServiceName: service, Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GET /subscriptions/summary?from=MM-YYYY&to=MM-YYYY&user_id=&service_name=
// Summary of subscriptions cost
// @Summary      Sum subscriptions cost for a period
// @Description  Считает сумму (в рублях) по активным месяцам в интервале [from,to] с фильтрами
// @Tags         subscriptions
// @Produce      json
// @Param        from          query     string  true   "Начало периода (MM-YYYY)"
// @Param        to            query     string  true   "Конец периода (MM-YYYY)"
// @Param        user_id       query     string  false  "UUID пользователя"
// @Param        service_name  query     string  false  "Название сервиса"
// @Success      200           {object}  map[string]int64  "Сумма, ключ total_rub"
// @Failure      400           {object}  map[string]string "Bad request"
// @Failure      500           {object}  map[string]string "Internal error"
// @Router       /subscriptions/summary [get]
func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromS, toS := q.Get("from"), q.Get("to")
	if fromS == "" || toS == "" {
		writeError(w, http.StatusBadRequest, "from/to required MM-YYYY")
		return
	}
	from, err := parseMonthYear(fromS)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad from")
		return
	}
	to, err := parseMonthYear(toS)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad to")
		return
	}
	if to.Before(from) {
		writeError(w, http.StatusBadRequest, "to before from")
		return
	}

	var uid *uuid.UUID
	if s := strings.TrimSpace(q.Get("user_id")); s != "" {
		u, err := uuid.Parse(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad user_id")
			return
		}
		uid = &u
	}
	var service *string
	if s := strings.TrimSpace(q.Get("service_name")); s != "" {
		service = &s
	}

	total, err := h.Repo.Summary(r.Context(), monthStart(from), monthStart(to), uid, service)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"total_rub": total})
}

// helpers

func parseMonthYear(s string) (time.Time, error) {
	// expected MM-YYYY
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return time.Time{}, errors.New("bad format")
	}
	m, err := atoi(parts[0])
	if err != nil || m < 1 || m > 12 {
		return time.Time{}, errors.New("bad month")
	}
	y, err := atoi(parts[1])
	if err != nil || y < 1 {
		return time.Time{}, errors.New("bad year")
	}
	return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC), nil
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func atoi(s string) (int, error) { var i int; _, err := fmt.Sscanf(s, "%d", &i); return i, err }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
