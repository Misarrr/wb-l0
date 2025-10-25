package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	log.Println("Запуск publisher...")

	// Чтение model.json
	data, err := os.ReadFile("model.json")
	if err != nil {
		log.Fatal("Ошибка чтения model.json:", err)
	}

	// Проверка валидности JSON
	var testOrder Order
	if err := json.Unmarshal(data, &testOrder); err != nil {
		log.Fatal("Некорректный JSON в model.json:", err)
	}

	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("Ошибка подключения к NATS:", err)
	}
	defer sc.Close()

	log.Println("Подключено к NATS Streaming")

	// Публикация сообщения
	err = sc.Publish("orders", data)
	if err != nil {
		log.Fatal("Ошибка публикации:", err)
	}

	log.Printf("Сообщение опубликовано: %s", testOrder.OrderUID)

	// Даем время на доставку
	time.Sleep(1 * time.Second)
	log.Println("Готово!")
}
