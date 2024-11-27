package handlers

import (
	"net/http"

	"api.hloud.fprzg.net/internal/info"
	"api.hloud.fprzg.net/internal/utils"
	"github.com/julienschmidt/httprouter"
)

type HealthCheckHandlers struct {
	cfg   *info.Config
	build *info.Build
}

func (h *HealthCheckHandlers) Routes(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", h.HealthCheck)
}

func (h *HealthCheckHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	env := utils.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": h.cfg.Env,
			"version":     h.build.Version,
		},
	}

	err := utils.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		internalErrorResponse(w, err)
	}
}
