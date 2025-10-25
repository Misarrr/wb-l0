package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/stan.go"
)

const (
	clusterID   = "test-cluster"
	clientID    = "order-service"
	channelName = "orders"
)

// getEnv - получение переменной окружения с дефолтным значением
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	log.Println("Запуск сервиса...")

	// Подключение к PostgreSQL
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "wbuser")
	dbPassword := getEnv("DB_PASSWORD", "wbpass")
	dbName := getEnv("DB_NAME", "orders_db")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := NewDB(connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	log.Println("БД подключена")

	// Создание кэша
	cache := NewCache()

	// Восстановление кэша из БД
	log.Println("Восстановление кэша из БД...")
	orders, err := db.GetAllOrders()
	if err != nil {
		log.Println("Ошибка восстановления кэша:", err)
	} else {
		for _, order := range orders {
			cache.Set(order.OrderUID, order)
		}
		log.Printf("Кэш восстановлен: %d заказов", cache.Count())
	}

	// Подключение к NATS Streaming
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))

	if err != nil {
		log.Fatal("Ошибка подключения к NATS:", err)
	}
	defer sc.Close()
	log.Println("NATS Streaming подключен")

	// Подписка на канал
	_, err = sc.Subscribe(channelName, func(m *stan.Msg) {
		log.Printf("Получено сообщение: %s", string(m.Data))

		// Валидация данных
		order, err := ValidateOrder(m.Data)
		if err != nil {
			log.Printf("Ошибка валидации: %v", err)
			return
		}

		// Сохранение в БД
		if err := db.SaveOrder(order); err != nil {
			log.Printf("Ошибка сохранения в БД: %v", err)
			return
		}

		// Добавление в кэш
		cache.Set(order.OrderUID, order)
		log.Printf("Заказ %s успешно обработан", order.OrderUID)
	}, stan.DurableName("order-service-durable"))

	if err != nil {
		log.Fatal("Ошибка подписки на канал:", err)
	}
	log.Printf("Подписка на канал '%s' активна", channelName)

	// HTTP сервер
	handler := NewHandler(cache)
	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/api/order", handler.GetOrderHandler)

	go func() {
		log.Println("HTTP сервер запущен на http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Ошибка запуска HTTP сервера:", err)
		}
	}()

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Завершение работы сервиса...")
}
