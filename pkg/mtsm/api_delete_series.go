package mtsm

import (
	"fmt"
)

type api_delete_series struct {
	// http 请求状态码
	HttpReqCode int `json:"http_req_code"`
	client      *Client
}

func (t *api_delete_series) request(body string) error {
	url := "/prometheus/api/v1/admin/tsdb/delete_series"
	resp, err := t.client.config.resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetResult(t).
		SetBody(body).
		Post(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	t.HttpReqCode = resp.StatusCode()

	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
