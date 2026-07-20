package workers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *Service) sendPostRequest(ctx context.Context, url string, payload json.RawMessage, secret []byte) (_ int, _ string, err error) {
	hash := hmac.New(sha256.New, secret)
	hash.Write(payload)
	signature := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err.Error(), err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err.Error(), err
	}

	defer func(body io.ReadCloser) {
		if err == nil {
			err = body.Close()
			return
		}

		_ = body.Close()
	}(resp.Body)

	return resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode), nil
}

func isSuccessCode(httpCode int) bool {
	return httpCode >= 200 && httpCode < 300
}
