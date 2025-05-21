package http

import (
	"net/http"

	"remote-patient-monitoring-system/internal/application/ingest"

	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	Service *ingest.IngestService
}

func NewIngestHandler(svc *ingest.IngestService) *IngestHandler {
	return &IngestHandler{Service: svc}
}

func (h *IngestHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/observations", h.postObservation)
}

func (h *IngestHandler) postObservation(c *gin.Context) {
	var input ingest.TelemetryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.Execute(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusAccepted)
}
