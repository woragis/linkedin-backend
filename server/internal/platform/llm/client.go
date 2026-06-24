package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	model   string
	timeout time.Duration
	http    *http.Client
}

func NewFromEnv() *Client {
	timeout := 30 * time.Second
	return &Client{
		apiKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		model:   envOr("LLM_MODEL", "gpt-4o-mini"),
		timeout: timeout,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("OPENAI_API_KEY not configured")
	}
	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.85,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai status %d: %s", res.StatusCode, string(body))
	}
	var out chatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	text := strings.TrimSpace(out.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("openai returned empty content")
	}
	if len(text) > 500 {
		text = text[:500]
	}
	return text, nil
}

const systemPromptPT = "Você escreve posts e comentários curtos para uma rede social profissional brasileira (estilo LinkedIn). Responda só com o texto final em português do Brasil, sem aspas, sem markdown, máximo 280 caracteres."

func (c *Client) GeneratePost(ctx context.Context, fullName, headline, topic string) (string, error) {
	prompt := fmt.Sprintf(
		"Escreva um post autêntico para %s (%s). Tema: %s. Tom natural, 1-3 frases.",
		fullName, headline, topic,
	)
	return c.Complete(ctx, systemPromptPT, prompt, 180)
}

func (c *Client) GenerateComment(ctx context.Context, fullName, headline, postBody, parentComment string) (string, error) {
	var prompt string
	if parentComment != "" {
		prompt = fmt.Sprintf(
			"%s (%s) responde em thread. Post: %s. Comentário pai: %s. Resposta curta.",
			fullName, headline, truncate(postBody, 200), truncate(parentComment, 200),
		)
	} else {
		prompt = fmt.Sprintf(
			"%s (%s) comenta um post: %s. Comentário curto e relevante.",
			fullName, headline, truncate(postBody, 240),
		)
	}
	return c.Complete(ctx, systemPromptPT, prompt, 120)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

var templateMarkers = []string{
	"Migrando mais um serviço para Go",
	"Alguém mais usando Redis como fila",
	"Treino de pernas feito",
	"Buscando estágio em tecnologia",
	"Concordo — já vi isso em produção",
}

func IsTemplateText(text string) bool {
	for _, m := range templateMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func LooksPortuguese(text string) bool {
	lower := strings.ToLower(text)
	for _, w := range []string{" de ", " em ", " que ", " para ", " com ", "ção", "ões"} {
		if strings.Contains(lower, w) {
			return true
		}
	}
	for _, r := range text {
		if strings.ContainsRune("áàâãéêíóôõúç", r) {
			return true
		}
	}
	return len(text) >= 24
}
