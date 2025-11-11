package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TelegramSender struct {
	botToken string
	client   *http.Client
}

func NewTelegramSender(botToken string) *TelegramSender {
	return &TelegramSender{
		botToken: botToken,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramSender) Name() string { return "telegram" }

func (t *TelegramSender) Send(ctx context.Context, recipient, message string) error {
	if t.botToken == "" {
		return errors.New("telegram bot token empty")
	}
	chat := recipient
	if chat == "" {
		return errors.New("telegram chat id empty")
	}
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	data := url.Values{}
	data.Set("chat_id", chat)
	data.Set("text", message)

	req, err := http.NewRequestWithContext(ctx, "POST", api, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if !r.Ok {
		return errors.New("telegram send failed: " + r.Description)
	}
	return nil
}
