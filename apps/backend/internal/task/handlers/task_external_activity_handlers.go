package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
	"github.com/kandev/kandev/internal/task/service"
)

const externalActivityBodyLimit = 16 * 1024

func (h *TaskHandlers) httpRecordExternalActivity(c *gin.Context) {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error_code": "json_required"})
		return
	}
	reader := http.MaxBytesReader(c.Writer, c.Request.Body, externalActivityBodyLimit)
	defer func() { _ = reader.Close() }()
	body, err := io.ReadAll(reader)
	if err != nil {
		status := http.StatusBadRequest
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		c.JSON(status, gin.H{"error_code": "invalid_external_activity"})
		return
	}
	activity, err := decodeExternalActivity(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "invalid_external_activity"})
		return
	}
	receipt, err := h.service.RecordExternalActivity(c.Request.Context(), c.Param("id"), activity)
	if err != nil {
		externalActivityError(c, err)
		return
	}
	status := http.StatusOK
	if receipt.Created {
		status = http.StatusCreated
	}
	c.JSON(status, receipt)
}

func decodeExternalActivity(body []byte) (*models.ExternalActivity, error) {
	if !utf8.Valid(body) {
		return nil, models.ErrInvalidExternalActivity
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var activity models.ExternalActivity
	if err := decoder.Decode(&activity); err != nil {
		return nil, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, models.ErrInvalidExternalActivity
	}
	return &activity, nil
}

func externalActivityError(c *gin.Context, err error) {
	if isClientDisconnect(err) {
		abortClientDisconnect(c)
		return
	}
	status, code := http.StatusServiceUnavailable, "external_activity_unavailable"
	switch {
	case isNotFound(err):
		status, code = http.StatusNotFound, "task_not_found"
	case service.IsForbidden(err):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, models.ErrInvalidExternalActivity):
		status, code = http.StatusBadRequest, "invalid_external_activity"
	case errors.Is(err, repoerrors.ErrExternalActivityConflict):
		status, code = http.StatusConflict, "external_activity_conflict"
	}
	c.JSON(status, gin.H{"error_code": code})
}
