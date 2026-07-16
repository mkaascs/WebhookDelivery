package workers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

func sendPostRequest(url string, payload json.RawMessage, secret []byte) (int, error) {
	hash := hmac.New(sha256.New, secret)
	hash.Write(payload)
	signature := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	statusCode := resp.StatusCode
	return statusCode, resp.Body.Close()
}

func isSuccessCode(httpCode int) bool {
	return httpCode >= 200 && httpCode < 300
}
