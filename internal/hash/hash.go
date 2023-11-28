package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/pavelborisofff/go-metrics/internal/config"
	"io"
	"net/http"
)

var cfg, _ = config.GetServerConfig()

func Make(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func Verify(body []byte, key, expectedHash string) (string, bool) {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	bodyHash := hex.EncodeToString(h.Sum(nil))
	return bodyHash, bodyHash == expectedHash
}

func HashHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HashSHA256") == "" || !cfg.UseHashKey {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "Error read body", http.StatusBadRequest)
			return
		}

		hashed, ok := Verify(body, cfg.HashKey, r.Header.Get("HashSHA256"))
		if !ok {
			http.Error(w, "Invalid hash", http.StatusBadRequest)
			return
		}

		w.Header().Set("HashSHA256", hashed)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)
	})
}
