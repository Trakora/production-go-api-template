package health

import (
	"net/http"
	"production-go-api-template/pkg/router"
	"time"

	"github.com/joeshaw/envdecode"
)

type Health struct {
	startTime time.Time
}

func NewHealthHandler() *Health {
	return &Health{startTime: time.Now()}
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	router.RespondWithJSON(r, w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) CheckHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now()
	uptime := currentTime.Sub(h.startTime)

	var healthcheck Check
	if err := envdecode.Decode(&healthcheck); err != nil {
		router.RespondWithError(r, w, http.StatusInternalServerError, "Error decoding environment variables", err)
		return
	}

	healthcheck.Status = "healthy"
	healthcheck.Timestamp = currentTime
	healthcheck.Uptime = uptime

	router.RespondWithJSON(r, w, http.StatusOK, healthcheck)
}
