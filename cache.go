package main

import (
	"log"
	"sync"
)

// Cache - кэш для хранения заказов в памяти
type Cache struct {
	data map[string]*Order
	mu   sync.RWMutex
}

// NewCache - создание нового кэша
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*Order),
	}
}

// Set - добавить заказ в кэш
func (c *Cache) Set(orderUID string, order *Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[orderUID] = order
	log.Printf("Заказ %s добавлен в кэш", orderUID)
}

// Get - получить заказ из кэша
func (c *Cache) Get(orderUID string) (*Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.data[orderUID]
	return order, exists
}

// GetAll - получить все заказы из кэша
func (c *Cache) GetAll() []*Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]*Order, 0, len(c.data))
	for _, order := range c.data {
		orders = append(orders, order)
	}
	return orders
}

// Count - количество заказов в кэше
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
