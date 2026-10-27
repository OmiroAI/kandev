package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kandev/kandev/internal/analytics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kandev/kandev/internal/analytics/models"
	"github.com/kandev/kandev/internal/analytics/repository"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
)

type usageStatsHTTPRepo struct {
	repository.Repository
	query models.UsageStatsQuery
	calls int
	err   error
}

func (r *usageStatsHTTPRepo) GetUsageStats(_ context.Context, _ string, query models.UsageStatsQuery) (*models.UsageStats, error) {
	r.calls++
	r.query = query
	if r.err != nil {
		return nil, r.err
	}
	known := int64(12)
	return &models.UsageStats{Totals: models.UsageTotals{KnownTokensTotal: known, TokensTotal: &known}}, nil
}

type usageStatsAuthorizer struct{ allowed bool }

func (a usageStatsAuthorizer) AuthorizeWorkspaceAccess(context.Context, string) error {
	if !a.allowed {
		return repoerrors.ErrWorkspaceNotFound
	}
	return nil
}

func TestUsageStatsHTTPAuthorizesBeforeParsingOrReading(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &usageStatsHTTPRepo{}
	h := NewStatsHandlers(repo, usageStatsAuthorizer{}, authzTestLogger(t))
	router := gin.New()
	h.registerHTTP(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/foreign/stats/usage?source=invalid", nil))
	if response.Code != http.StatusNotFound || repo.calls != 0 {
		t.Fatalf("status=%d reads=%d", response.Code, repo.calls)
	}
}

func TestUsageStatsHTTPMapsSourceRangesAndPages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &usageStatsHTTPRepo{}
	h := NewStatsHandlers(repo, usageStatsAuthorizer{allowed: true}, authzTestLogger(t))
	router := gin.New()
	h.registerHTTP(router)
	for _, invalid := range []string{"source=invalid", "limit=0", "limit=101", "range=year"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws/stats/usage?"+invalid, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("query=%s status=%d", invalid, response.Code)
		}
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws/stats/usage?range=all&source=external&limit=12&task_cursor=task-a", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if repo.query.Start != nil || repo.query.Source != "external" || repo.query.Limit != 12 || repo.query.TaskCursor != "task-a" {
		t.Fatalf("query=%+v", repo.query)
	}
	if days := repo.query.Until.Sub(repo.query.DailyStart).Hours() / 24; days != 365 {
		t.Fatalf("daily horizon=%v", days)
	}
	var stats models.UsageStats
	if err := json.Unmarshal(response.Body.Bytes(), &stats); err != nil || stats.Totals.TokensTotal == nil || *stats.Totals.TokensTotal != 12 {
		t.Fatalf("response=%+v error=%v", stats, err)
	}
}

func (r *recordingRepo) GetUsageStats(_ context.Context, _ string, _ models.UsageStatsQuery) (*models.UsageStats, error) {
	r.calls = append(r.calls, "GetUsageStats")
	return &models.UsageStats{}, nil
}

func TestUsageStatsHTTPUnavailableAndSafeErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		repo   repository.Repository
		status int
		code   string
	}{
		{"unwired", struct{ repository.Repository }{}, http.StatusServiceUnavailable, "usage_statistics_unavailable"},
		{"invalid cursor", &usageStatsHTTPRepo{err: analytics.ErrInvalidUsageQuery}, http.StatusBadRequest, "invalid_usage_query"},
		{"busy", &usageStatsHTTPRepo{err: analytics.ErrAnalyticsBusy}, http.StatusServiceUnavailable, "analytics_busy"},
		{"database failure", &usageStatsHTTPRepo{err: errors.New("private-database-path")}, http.StatusInternalServerError, "failed to get usage statistics"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			NewStatsHandlers(test.repo, usageStatsAuthorizer{allowed: true}, authzTestLogger(t)).registerHTTP(router)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws/stats/usage", nil))
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["error_code"] != test.code && body["error"] != test.code {
				t.Fatalf("unsafe body=%s", response.Body.String())
			}
		})
	}
}
