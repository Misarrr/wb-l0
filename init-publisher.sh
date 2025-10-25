#!/bin/bash
echo "Ожидание запуска NATS..."
sleep 10

echo "Публикация тестового заказа..."
go run /app/publisher.go /app/models.go

echo "Готово! Данные опубликованы."
