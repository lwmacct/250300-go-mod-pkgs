package mtsm

import (
	"fmt"
)

type api_label_values struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
	client *Client
}

func (t *api_label_values) request(labelName string, params map[string]string) error {
	url := fmt.Sprintf("/prometheus/api/v1/label/%s/values", labelName)
	resp, err := t.client.config.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(t).
		SetQueryParams(params).
		Get(url)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
