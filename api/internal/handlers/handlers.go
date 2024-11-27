package handlers

import (
	"api.hloud.fprzg.net/internal/info"
	"github.com/julienschmidt/httprouter"
)

type RouteHandlers interface {
	Routes(httprouter.Router) httprouter.Router
}

type Handlers struct {
	HealthCheck HealthCheckHandlers
	Files       FileHandlers
}

func NewHandlers(cfg *info.Config, build *info.Build) Handlers {
	return Handlers{
		HealthCheck: HealthCheckHandlers{cfg: cfg, build: build},
		Files:       FileHandlers{cfg: cfg, build: build},
	}
}
