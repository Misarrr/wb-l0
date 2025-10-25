package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Настройки стресс-теста
	totalRequests := 10000 // Всего запросов
	concurrentUsers := 100 // Одновременных пользователей
	url := "http://localhost:8080/api/order?id=b563feb7b2b84b6test"

	fmt.Printf("🚀 Запуск стресс-теста\n")
	fmt.Printf("   Всего запросов: %d\n", totalRequests)
	fmt.Printf("   Параллельных: %d\n", concurrentUsers)
	fmt.Printf("   URL: %s\n\n", url)

	// Счётчики
	var (
		successCount int
		errorCount   int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	startTime := time.Now()

	// Канал для распределения работы
	jobs := make(chan int, totalRequests)
	for i := 0; i < totalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	// Запуск воркеров
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				resp, err := http.Get(url)
				mu.Lock()
				if err != nil || resp.StatusCode != 200 {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Результаты
	fmt.Printf("\n📊 РЕЗУЛЬТАТЫ:\n")
	fmt.Printf("   ✅ Успешных: %d\n", successCount)
	fmt.Printf("   ❌ Ошибок: %d\n", errorCount)
	fmt.Printf("   ⏱️  Время: %v\n", duration)
	fmt.Printf("   📈 RPS (запросов/сек): %.2f\n", float64(totalRequests)/duration.Seconds())
	fmt.Printf("   ⚡ Среднее время ответа: %v\n", duration/time.Duration(totalRequests))
}
