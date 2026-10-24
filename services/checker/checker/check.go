package checker

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/mbrysin/PingTower/services/checker/model"
)

func Check(ctx context.Context, rawURL string) model.CheckResult {
	result := model.CheckResult{
		CheckedAt: time.Now().UTC(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	req.Header.Set("User-Agent", "PingTower/0.0-alpha")

	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 32*1024))

	result.StatusCode = resp.StatusCode
	result.IsUp = resp.StatusCode < http.StatusInternalServerError

	return result
}
