package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kandev/kandev/internal/analytics"
	"github.com/kandev/kandev/internal/analytics/models"
	"github.com/kandev/kandev/internal/analytics/repository"
)

func (h *StatsHandlers) httpGetUsageStats(c *gin.Context) {
	workspaceID, _, days, ok := h.parseRequest(c)
	if !ok {
		return
	}
	query, err := parseUsageStatsQuery(c, days)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "invalid_usage_query"})
		return
	}
	reader, ok := h.repo.(repository.UsageStatsReader)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error_code": "usage_statistics_unavailable"})
		return
	}
	stats, err := reader.GetUsageStats(c.Request.Context(), workspaceID, query)
	if errors.Is(err, analytics.ErrInvalidUsageQuery) {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "invalid_usage_query"})
		return
	}
	if err != nil {
		h.fail(c, workspaceID, "usage statistics", err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func parseUsageStatsQuery(c *gin.Context, days int) (models.UsageStatsQuery, error) {
	query := models.UsageStatsQuery{
		Source: c.DefaultQuery("source", "all"), Limit: 50,
		TaskCursor: c.Query("task_cursor"), ProviderCursor: c.Query("provider_cursor"),
	}
	switch c.Query("range") {
	case "", "week", "month", "all":
	default:
		return query, analytics.ErrInvalidUsageQuery
	}
	if query.Source != "all" && query.Source != "native" && query.Source != "external" {
		return query, analytics.ErrInvalidUsageQuery
	}
	if value, present := c.GetQuery("limit"); present {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			return query, analytics.ErrInvalidUsageQuery
		}
		query.Limit = limit
	}
	query.Until = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
	query.DailyStart = query.Until.AddDate(0, 0, -days)
	if c.Query("range") != "all" {
		query.Start = &query.DailyStart
	}
	return query, nil
}
