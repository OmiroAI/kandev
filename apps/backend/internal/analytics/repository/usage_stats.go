package repository

import (
	"context"

	"github.com/kandev/kandev/internal/analytics/models"
)

// UsageStatsReader projects task accounting independently of native session state.
type UsageStatsReader interface {
	GetUsageStats(context.Context, string, models.UsageStatsQuery) (*models.UsageStats, error)
}
