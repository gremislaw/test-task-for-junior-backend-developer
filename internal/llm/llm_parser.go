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

type Parser struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func New() *Parser {
	return &Parser{
		apiKey:  os.Getenv("LLM_API_KEY"),
		baseURL: os.Getenv("LLM_BASE_URL"),
		model:   os.Getenv("LLM_MODEL"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Parser) Valid() bool {
	return p.apiKey != "" && p.baseURL != ""
}

type TaskExtraction struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	DueDate        string `json:"due_date"`
	RecurrenceType string `json:"recurrence_type"`
	IntervalDays   *int   `json:"interval_days,omitempty"`
	DaysOfMonth    []int  `json:"days_of_month,omitempty"`
	Parity         string `json:"parity,omitempty"`
}

func (p *Parser) Parse(ctx context.Context, text string) (*TaskExtraction, error) {
	if !p.Valid() {
		return nil, ErrLLMNotConfigured
	}

	t := time.Now()
	s := fmt.Sprintf("Сегодня %d-%02d-%02d %02d:%02d:%02d",
    t.Year(), t.Month(), t.Day(),
    t.Hour(), t.Minute(), t.Second())

	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": s},
			{
				"role": "system",
				"content": `Extract task details from Russian free text. Return STRICT JSON:
{"title": "short actionable title", "description": "full context or details", "due_date": "RFC3339", "recurrence_type": "daily|monthly|even_odd|none", "interval_days": N, "days_of_month": [1,15], "parity": "even|odd"}
Examples:
- "обойти Иванова каждые 3 дня" → {"title": "Обход Иванова", "description": "", "status": "new" "due_date":null, "is_recurrence": true, "recurrence_type":"daily", "interval_days":3, "days_of_month": [], "parity": ""}
- "визит 10 мая в 14:00" → {"title": "визит", "description": "", "due_date":"2026-05-10T14:00:00Z", "is_recurrence": false, "recurrence_type":"none"}
- "проверять по нечётным числам и давать анальгин" → {"title": "Проверка", "description": "Дать анальгин", "due_date":"", "is_recurrence": true, "recurrence_type":"even_odd", "interval_days":0, "days_of_month": [], "parity": "odd"}
RULES:
- Title: max 50 chars, imperative mood (e.g., "Обход Иванова"). If u see abbreviatures like "МРТ" dont change it just copy to title
- Description: more than 50 chars. its about what doctor will do in details (e.g., "Дать лекарство", "Поменять белье", "Вколоть обезбол")
- Due date: when it should be done, maybe its time when recurrence starts, parse if explicit ("завтра в 14:00", "после 10 мая 2026"), else "".
- Recurrence: extract ONLY if explicit.
	1. daily -> change interval_days (e.g. "каждые 3 дня", "еженедельно", "ежедневно", "каждый день", "каждые 2 недели")
	2. monthly -> change days_of_month (e.g. "каждый 3, 7 день месяца", "каждого 1 числа", "раз в месяц"). if days of month not given just use today
	3. even_odd -> change parity (e.g. "каждый четный день", "в нечетные дни", "по четным числам")
- NO markdown, NO explanations, ONLY valid JSON.`,
			},
			{"role": "user", "content": text},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	reqURL := strings.TrimRight(p.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(raw))
	}

	var apiResp struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	var ext TaskExtraction
	if err := json.Unmarshal([]byte(content), &ext); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return &ext, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
