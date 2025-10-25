package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB - обёртка для работы с базой данных
type DB struct {
	conn *sql.DB
}

// NewDB - создание подключения к БД
func NewDB(connStr string) (*DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("БД недоступна: %w", err)
	}

	log.Println("Подключение к PostgreSQL успешно")
	return &DB{conn: conn}, nil
}

// SaveOrder - сохранение заказа в БД (JSONB)
func (db *DB) SaveOrder(order *Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	_, err = db.conn.Exec(`
		INSERT INTO orders (order_uid, data) 
		VALUES ($1, $2)
		ON CONFLICT (order_uid) DO NOTHING
	`, order.OrderUID, orderJSON)

	return err
}

// GetAllOrders - получение всех заказов из БД (JSONB)
func (db *DB) GetAllOrders() ([]*Order, error) {
	rows, err := db.conn.Query(`SELECT order_uid, data FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*Order, 0)
	for rows.Next() {
		var orderUID string
		var orderData []byte

		if err := rows.Scan(&orderUID, &orderData); err != nil {
			return nil, err
		}

		var order Order
		if err := json.Unmarshal(orderData, &order); err != nil {
			log.Printf("Ошибка десериализации заказа %s: %v", orderUID, err)
			continue
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

// ValidateOrder - проверка корректности JSON заказа
func ValidateOrder(data []byte) (*Order, error) {
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("некорректный JSON: %w", err)
	}

	// Базовая валидация
	if order.OrderUID == "" {
		return nil, fmt.Errorf("отсутствует order_uid")
	}

	return &order, nil
}
