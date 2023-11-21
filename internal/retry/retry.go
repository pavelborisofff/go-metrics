package retry

import (
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"go.uber.org/zap"

	"github.com/pavelborisofff/go-metrics/internal/logger"
)

func Request(c *http.Client, w *http.Request) (*http.Response, error) {
	var r *http.Response
	var delay uint = 1
	log := logger.GetLogger()

	err := retry.Do(
		func() error {
			var err error
			r, err = c.Do(w)
			if err != nil {
				log.Error("Error do request, retry")
			} else {
				defer r.Body.Close()
			}
			return err
		},
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			curDelay := delay + 2*n
			time.Sleep(time.Duration(curDelay) * time.Second)
			log.Info("Retry request", zap.Uint("attempt", n), zap.Uint("delay", curDelay))
		}),
	)

	if err != nil {
		log.Error("Error retry request", zap.Error(err))
		return nil, err
	}

	return r, nil
}
