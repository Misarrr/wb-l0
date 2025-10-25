package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler - обработчик HTTP запросов
type Handler struct {
	cache *Cache
}

// NewHandler - создание нового обработчика
func NewHandler(cache *Cache) *Handler {
	return &Handler{cache: cache}
}

// GetOrderHandler - получение заказа по ID
func (h *Handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderUID := r.URL.Query().Get("id")
	if orderUID == "" {
		http.Error(w, "Параметр id обязателен", http.StatusBadRequest)
		return
	}

	order, exists := h.cache.Get(orderUID)
	if !exists {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
	log.Printf("Заказ %s отправлен клиенту", orderUID)
}

// IndexHandler - главная страница
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
