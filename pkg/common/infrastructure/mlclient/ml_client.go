package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Observation es el payload que enviamos al servicio ML
type MLObservation struct {
	ID                string  `json:"id"`
	PatientID         string  `json:"patient_id"`
	HeartRate         float64 `json:"heart_rate"`
	RespRate          float64 `json:"resp_rate"`
	Spo2              float64 `json:"spo2"`
	EffectiveDateTime string  `json:"effective_date_time"`
}

// MLAnomalyResult es la respuesta que esperamos del servicio ML
type MLAnomalyResult struct {
	Prediction   bool    `json:"prediction"`
	AnomalyScore float64 `json:"anomaly_score"`
}

// Client representa al cliente HTTP del servicio ML
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient crea un nuevo cliente para el servicio ML
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Predict(obs MLObservation) (*MLAnomalyResult, error) {
	url := fmt.Sprintf("%s/predict", c.baseURL)

	body, err := json.Marshal(obs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal observation: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create ML request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ML service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned non-OK status: %s", resp.Status)
	}

	var result MLAnomalyResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ML response: %w", err)
	}

	return &result, nil
}
