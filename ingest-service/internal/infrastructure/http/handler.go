package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lioarce01/remote-patient-monitoring-system/ingest-service/internal/application"
)

type IngestHandler struct {
	Service *application.IngestService
}

func NewIngestHandler(svc *application.IngestService) *IngestHandler {
	return &IngestHandler{Service: svc}
}

func (h *IngestHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/observations", h.postObservation)
}

func (h *IngestHandler) postObservation(c *gin.Context) {
	var input application.TelemetryInput
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
