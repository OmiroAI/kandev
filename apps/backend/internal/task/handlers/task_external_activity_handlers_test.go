package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/kandev/kandev/internal/auth/authn"
	dbutil "github.com/kandev/kandev/internal/db"
	"github.com/kandev/kandev/internal/task/models"
	tasksqlite "github.com/kandev/kandev/internal/task/repository/sqlite"
	"github.com/kandev/kandev/internal/task/service"
)

const externalActivityJSON = `{"schema_version":1,"source":"kodi","source_instance":"d2d93284-5aab-4384-b50c-ac5bebc8e133","event_id":"usage:1","kind":"worker_run","phase":"coding","outcome":"succeeded","occurred_at":"2026-10-25T00:00:00Z","input_tokens":0}`

func externalActivityHTTPFixture(t *testing.T) (*gin.Engine, *sqlx.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	raw, err := dbutil.OpenSQLite(filepath.Join(t.TempDir(), "external.db"))
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(raw, "sqlite3")
	t.Cleanup(func() { _ = db.Close() })
	repo, err := tasksqlite.NewWithDB(db, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repo.CreateWorkspace(ctx, &models.Workspace{ID: "external-ws", Name: "External", OwnerID: "owner"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTask(ctx, &models.Task{ID: "external-task", WorkspaceID: "external-ws", Title: "Example"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertWorkspaceMember(ctx, &models.WorkspaceMember{WorkspaceID: "external-ws", UserID: "viewer", Role: "viewer"}); err != nil {
		t.Fatal(err)
	}
	log := newTestLogger(t)
	svc := service.NewService(service.Repos{Workspaces: repo, Tasks: repo, Usage: repo}, nil, log, service.RepositoryDiscoveryConfig{})
	handler := &TaskHandlers{service: svc, logger: log}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		authn.SetOnGin(c, authn.Identity{UserID: c.GetHeader("X-Test-User"), Role: authn.RoleMember})
	})
	handler.registerHTTP(router)
	return router, db
}

func externalActivityHTTPRequest(router *gin.Engine, user, task, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+task+"/external-activity", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-User", user)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

// Covers AC-COSTS-EXTERNAL-001.1 and AC-COSTS-EXTERNAL-001.3.

func TestExternalActivityHTTPRejectsInvalidAndForbiddenWrites(t *testing.T) {
	router, db := externalActivityHTTPFixture(t)
	for _, test := range []struct {
		name, user, task, body string
		status                 int
	}{
		{"missing body", "owner", "external-task", "", http.StatusBadRequest},
		{"unknown field", "owner", "external-task", strings.TrimSuffix(externalActivityJSON, "}") + `,"prompt":"forbidden"}`, http.StatusBadRequest},
		{"trailing value", "owner", "external-task", externalActivityJSON + `{}`, http.StatusBadRequest},
		{"fractional tokens", "owner", "external-task", strings.Replace(externalActivityJSON, `"input_tokens":0`, `"input_tokens":1.5`, 1), http.StatusBadRequest},
		{"oversized body", "owner", "external-task", externalActivityJSON + strings.Repeat(" ", 16384), http.StatusRequestEntityTooLarge},
		{"invalid UTF8", "owner", "external-task", externalActivityJSON + string([]byte{255}), http.StatusBadRequest},
		{"read only", "viewer", "external-task", externalActivityJSON, http.StatusForbidden},
		{"foreign", "foreign", "external-task", externalActivityJSON, http.StatusNotFound},
		{"missing task", "owner", "missing", externalActivityJSON, http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := externalActivityHTTPRequest(router, test.user, test.task, test.body)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM task_external_activity`); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rejected requests wrote %d observations", count)
	}
}

func TestExternalActivityHTTPInsertReplayAndConflict(t *testing.T) {
	router, db := externalActivityHTTPFixture(t)
	first := externalActivityHTTPRequest(router, "owner", "external-task", externalActivityJSON)
	if first.Code != http.StatusCreated {
		t.Fatalf("first status=%d body=%s", first.Code, first.Body.String())
	}
	var receipt models.ExternalActivityReceipt
	if err := json.Unmarshal(first.Body.Bytes(), &receipt); err != nil || receipt.ID == "" || !receipt.Created {
		t.Fatalf("receipt=%+v error=%v", receipt, err)
	}
	replay := externalActivityHTTPRequest(router, "owner", "external-task", externalActivityJSON)
	if replay.Code != http.StatusOK {
		t.Fatalf("replay status=%d", replay.Code)
	}
	var duplicate models.ExternalActivityReceipt
	if err := json.Unmarshal(replay.Body.Bytes(), &duplicate); err != nil || duplicate.ID != receipt.ID || duplicate.Created {
		t.Fatalf("duplicate=%+v error=%v", duplicate, err)
	}
	conflict := externalActivityHTTPRequest(router, "owner", "external-task", strings.Replace(externalActivityJSON, `"input_tokens":0`, `"input_tokens":1`, 1))
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status=%d", conflict.Code)
	}
	var native int
	if err := db.Get(&native, `SELECT COUNT(*) FROM task_sessions`); err != nil {
		t.Fatal(err)
	}
	if native != 0 {
		t.Fatal("ingestion started a native session")
	}
}

func TestExternalActivityHTTPContentTypeAndStorageFailure(t *testing.T) {
	router, db := externalActivityHTTPFixture(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/external-task/external-activity", strings.NewReader(externalActivityJSON))
	request.Header.Set("X-Test-User", "owner")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("content type status=%d", response.Code)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	response = externalActivityHTTPRequest(router, "owner", "external-task", externalActivityJSON)
	if response.Code != http.StatusServiceUnavailable || response.Body.String() != `{"error_code":"external_activity_unavailable"}` {
		t.Fatalf("storage failure status=%d body=%s", response.Code, response.Body.String())
	}
}
