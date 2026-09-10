package main

import (
	"fmt"
	"log"
	stdhttp "net/http"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"gotracker/internal/cache"
	apihttp "gotracker/internal/http"
	"gotracker/internal/queue"
	"gotracker/internal/repository"
	"gotracker/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env не найден, используем системные переменные окружения")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("не удалось подключиться к PostgreSQL:", err)
	}
	defer db.Close()

	log.Println("Подключение к PostgreSQL установлено")

	cacheTTL, err := strconv.Atoi(
		os.Getenv("REDIS_TTL"),
	)

	if err != nil {
		cacheTTL = 60
	}

	redisCache := cache.NewRedisCache(
		os.Getenv("REDIS_ADDR"),
		cacheTTL,
	)

	queue.StartConsumer()

	repo := repository.NewPostgresOrderRepo(db)

	svc := service.NewOrderService(repo, redisCache)

	server := apihttp.NewServer(svc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf(
		"GoTracker API запущен на http://localhost:%s",
		port,
	)

	if err := stdhttp.ListenAndServe(":"+port, server); err != nil {
		log.Fatal(err)
	}

	fmt.Println(undefinedVariable)
}
