package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Service struct {
	redis  *redis.Client
	twilio *twilio.RestClient
}

func NewService(redisClient *redis.Client, accountSID, authToken string) *Service {
	return &Service{
		redis:  redisClient,
		twilio: twilio.NewRestClientWithParams(twilio.ClientParams{Username: accountSID, Password: authToken}),
	}
}

func (s *Service) SendOTP(ctx context.Context, phone string) error {
	// Генерируем 6-значный код
	code, err := generateOTP()
	if err != nil {
		return fmt.Errorf("ошибка генерации OTP: %w", err)
	}

	// Сохраняем в Redis с TTL 5 минут
	key := fmt.Sprintf("otp:%s", phone)
	err = s.redis.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("ошибка сохранения OTP в Redis: %w", err)
	}

	// Отправляем SMS через Twilio
	params := &openapi.CreateMessageParams{}
	params.SetTo(phone)
	params.SetBody(fmt.Sprintf("Ваш код подтверждения: %s", code))

	_, err = s.twilio.Api.CreateMessage(params)
	if err != nil {
		// Удаляем из Redis если SMS не отправился
		s.redis.Del(ctx, key)
		return fmt.Errorf("ошибка отправки SMS: %w", err)
	}

	log.Printf("OTP отправлен на номер %s", phone)
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, phone, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)

	// Получаем код из Redis
	storedCode, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Код не найден или истек
	}
	if err != nil {
		return false, fmt.Errorf("ошибка получения OTP из Redis: %w", err)
	}

	// Проверяем код
	if storedCode == code {
		// Удаляем код после успешной проверки
		s.redis.Del(ctx, key)
		log.Printf("OTP успешно верифицирован для номера %s", phone)
		return true, nil
	}

	return false, nil
}

func generateOTP() (string, error) {
	// Генерируем случайное число от 100000 до 999999
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
