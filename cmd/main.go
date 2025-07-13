package main

import (
	"FixDrive/config"
	"FixDrive/internal/admin"
	"FixDrive/internal/admin/handler"
	clientAuth "FixDrive/internal/client/auth"
	clientHandler "FixDrive/internal/client/auth/handler"
	driverAuth "FixDrive/internal/driver/auth"
	driverHandler "FixDrive/internal/driver/auth/handler"
	"FixDrive/internal/otp"
	otpHandler "FixDrive/internal/otp/handler"
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"

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

	// Проверяем подключение к БД
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка ping БД:", err)
	}
	log.Println("Подключение к базе данных успешно")

	// Подключаемся к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверяем подключение к Redis
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Ошибка подключения к Redis:", err)
	}
	log.Println("Подключение к Redis успешно")

	// Создаем сервисы авторизации
	clientService := clientAuth.NewService(db, cfg.JWT.Secret)
	clientHandlers := clientHandler.NewHandler(clientService)

	driverService := driverAuth.NewService(db, cfg.JWT.Secret)
	driverHandlers := driverHandler.NewHandler(driverService)

	// Создаем OTP сервис
	otpService := otp.NewService(rdb, cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.FromPhone)
	otpHandlers := otpHandler.NewHandler(otpService)

	// Настраиваем роутер
	r := chi.NewRouter()

	// Добавляем базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Монтируем роуты авторизации
	r.Mount("/auth/client", clientHandlers.Routes())
	r.Mount("/auth/driver", driverHandlers.Routes())

	// Монтируем OTP роуты
	r.Mount("/otp", otpHandlers.Routes())

	// Добавляем простой health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	userRepo := admin.NewUserRepo(db)
	h := handler.NewHandler(userRepo)

	r.Post("/loginAdmin", h.Login)

	log.Printf("Сервер запущен на порту %s", cfg.Server.Port)
	log.Println("Client auth: /auth/client/*")
	log.Println("Driver auth: /auth/driver/*")
	log.Println("OTP: /otp/*")

	// ВАЖНО: передаем роутер в ListenAndServe
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, r))
}
