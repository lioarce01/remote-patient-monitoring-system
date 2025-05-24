package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lioarce01/remote-patient-monitoring-system/api-service/internal/application"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/domain/entities"
)

type QueryHandler struct {
	Service *application.QueryService
}

func NewQueryHandler(svc *application.QueryService) *QueryHandler {
	return &QueryHandler{Service: svc}
}

func (h *QueryHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/patients/:id/observations", h.getObservations)
	r.GET("/patients/:id/alerts", h.getAlerts)
}

func (h *QueryHandler) getObservations(c *gin.Context) {
	id := c.Param("id")
	from := c.Query("from")
	to := c.Query("to")

	data, err := h.Service.GetPatientObservations(c.Request.Context(), id, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[getObservations] returning %d observations", len(data))
	for i, obs := range data {
		log.Printf("Observation %d: %+v", i, obs)
	}

	if len(data) == 0 {
		data = []entities.Observation{}
	}

	c.JSON(http.StatusOK, data)
}

func (h *QueryHandler) getAlerts(c *gin.Context) {
	id := c.Param("id")

	data, err := h.Service.GetPatientAlerts(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
