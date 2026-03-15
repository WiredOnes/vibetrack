package logic

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Константы API GigaChat
const (
	authURL     = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	gigachatURL = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	scope       = "GIGACHAT_API_PERS"
)

// GigaChatClient - клиент для работы с API GigaChat
type GigaChatClient struct {
	clientID     string
	clientSecret string
	authToken    string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

// AuthResponse - структура ответа авторизации
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// ChatRequest - структура запроса к GigaChat
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	UpdateAt int64     `json:"update_at,omitempty"`
}

// Message - структура сообщения
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse - структура ответа от GigaChat
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
}

// NewGigaChatClient создает новый клиент GigaChat
func NewGigaChatClient(clientID, clientSecret string) *GigaChatClient {
	return &GigaChatClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getAuthToken получает или обновляет токен авторизации
func (c *GigaChatClient) getAuthToken(ctx context.Context) error {
	// Проверяем, не истек ли текущий токен
	if c.authToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// Формируем запрос на авторизацию
	authData := fmt.Sprintf("scope=%s", scope)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBufferString(authData))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса авторизации: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(c.clientID, c.clientSecret))
	req.Header.Set("RqUID", generateRqUID())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса авторизации: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа авторизации: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка авторизации: статус %d, тело: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("ошибка парсинга ответа авторизации: %w", err)
	}

	c.authToken = authResp.AccessToken
	c.tokenExpiry = time.Unix(authResp.ExpiresAt, 0)

	return nil
}

// SendChatRequest отправляет запрос к GigaChat
func (c *GigaChatClient) SendChatRequest(ctx context.Context, messages []Message) (*ChatResponse, error) {
	// Получаем токен авторизации
	if err := c.getAuthToken(ctx); err != nil {
		return nil, err
	}

	// Формируем запрос
	chatReq := ChatRequest{
		Model:    "GigaChat",
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", gigachatURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: статус %d, тело: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &chatResp, nil
}

// basicAuth создает Basic Auth заголовок
func basicAuth(clientID, clientSecret string) string {
	auth := clientID + ":" + clientSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// generateRqUID генерирует уникальный идентификатор запроса
func generateRqUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
