package worker

import (
	"auction/auction/internal/service"
	"context"
	"log/slog"
	"time"
)

func RunExpiry(ctx context.Context, svc *service.Service, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := svc.Expire(ctx)
			if err != nil {
				logger.Error("expire auctions", "error", err)
			} else if count > 0 {
				logger.Info("expired auctions", "count", count)
			}
		}
	}
}
