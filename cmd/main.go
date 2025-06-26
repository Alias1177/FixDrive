package main

import (
	"FixDrive/config"
	clientAuth "FixDrive/internal/client/auth"
	clientHandler "FixDrive/internal/client/auth/handler"
	driverAuth "FixDrive/internal/driver/auth"
	driverHandler "FixDrive/internal/driver/auth/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка ping БД:", err)
	}

	log.Println("Подключение к базе данных успешно")

	// Создаем сервисы авторизации
	clientService := clientAuth.NewService(db, cfg.JWT.Secret)
	clientHandlers := clientHandler.NewHandler(clientService)

	driverService := driverAuth.NewService(db, cfg.JWT.Secret)
	driverHandlers := driverHandler.NewHandler(driverService)

	// Настраиваем роутер
	r := chi.NewRouter()

	// Добавляем базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Монтируем роуты авторизации
	r.Mount("/auth/client", clientHandlers.Routes())
	r.Mount("/auth/driver", driverHandlers.Routes())

	// Добавляем простой health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Сервер запущен на порту %s", cfg.Server.Port)
	log.Println("Client auth: /auth/client/*")
	log.Println("Driver auth: /auth/driver/*")

	// ВАЖНО: передаем роутер в ListenAndServe
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, r))
}
