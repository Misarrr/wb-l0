package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetOrderHandler_Success - тест успешного получения заказа
func TestGetOrderHandler_Success(t *testing.T) {
	// Создание тестового кэша
	cache := NewCache()
	testOrder := &Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		DateCreated: time.Now(),
		Delivery: Delivery{
			Name:  "Test User",
			Phone: "+1234567890",
			Email: "test@test.com",
		},
		Payment: Payment{
			Transaction: "test123",
			Amount:      1000,
			Currency:    "USD",
		},
		Items: []Item{
			{
				Name:  "Test Item",
				Price: 100,
			},
		},
	}
	cache.Set("test123", testOrder)

	// Создание handler
	handler := NewHandler(cache)

	// Создание запроса
	req := httptest.NewRequest("GET", "/api/order?id=test123", nil)
	w := httptest.NewRecorder()

	// Выполнение запроса
	handler.GetOrderHandler(w, req)

	// Проверка результата
	if w.Code != http.StatusOK {
		t.Errorf("Ожидался код 200, получен %d", w.Code)
	}

	var order Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if order.OrderUID != "test123" {
		t.Errorf("Ожидался OrderUID 'test123', получен '%s'", order.OrderUID)
	}
}

// TestGetOrderHandler_NotFound - тест несуществующего заказа
func TestGetOrderHandler_NotFound(t *testing.T) {
	cache := NewCache()
	handler := NewHandler(cache)

	req := httptest.NewRequest("GET", "/api/order?id=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.GetOrderHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался код 404, получен %d", w.Code)
	}
}

// TestGetOrderHandler_MissingID - тест без параметра ID
func TestGetOrderHandler_MissingID(t *testing.T) {
	cache := NewCache()
	handler := NewHandler(cache)

	req := httptest.NewRequest("GET", "/api/order", nil)
	w := httptest.NewRecorder()

	handler.GetOrderHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался код 400, получен %d", w.Code)
	}
}

// TestCache_SetGet - тест кэша
func TestCache_SetGet(t *testing.T) {
	cache := NewCache()

	order := &Order{
		OrderUID: "cache_test",
	}

	cache.Set("cache_test", order)

	retrieved, exists := cache.Get("cache_test")
	if !exists {
		t.Error("Заказ не найден в кэше")
	}

	if retrieved.OrderUID != "cache_test" {
		t.Errorf("Ожидался OrderUID 'cache_test', получен '%s'", retrieved.OrderUID)
	}
}

// TestCache_Count - тест подсчёта элементов в кэше
func TestCache_Count(t *testing.T) {
	cache := NewCache()

	if cache.Count() != 0 {
		t.Errorf("Ожидалось 0 элементов, получено %d", cache.Count())
	}

	cache.Set("test1", &Order{OrderUID: "test1"})
	cache.Set("test2", &Order{OrderUID: "test2"})

	if cache.Count() != 2 {
		t.Errorf("Ожидалось 2 элемента, получено %d", cache.Count())
	}
}

// TestValidateOrder_Valid - тест валидации корректного заказа
func TestValidateOrder_Valid(t *testing.T) {
	validJSON := `{
		"order_uid": "test123",
		"track_number": "TRACK123",
		"entry": "WBIL",
		"delivery": {},
		"payment": {},
		"items": []
	}`

	order, err := ValidateOrder([]byte(validJSON))
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}

	if order.OrderUID != "test123" {
		t.Errorf("Ожидался OrderUID 'test123', получен '%s'", order.OrderUID)
	}
}

// TestValidateOrder_Invalid - тест валидации некорректного JSON
func TestValidateOrder_Invalid(t *testing.T) {
	invalidJSON := `{invalid json`

	_, err := ValidateOrder([]byte(invalidJSON))
	if err == nil {
		t.Error("Ожидалась ошибка валидации")
	}
}

// TestValidateOrder_MissingOrderUID - тест валидации без order_uid
func TestValidateOrder_MissingOrderUID(t *testing.T) {
	jsonWithoutUID := `{
		"track_number": "TRACK123",
		"entry": "WBIL"
	}`

	_, err := ValidateOrder([]byte(jsonWithoutUID))
	if err == nil {
		t.Error("Ожидалась ошибка из-за отсутствия order_uid")
	}
}
