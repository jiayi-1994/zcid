package admin

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/pkg/database"
	"gorm.io/gorm"
)

type SystemSettings struct {
	K8sClusterURL   string `json:"k8sClusterUrl"`
	DefaultRegistry  string `json:"defaultRegistry"`
	GlobalSettings  map[string]string `json:"globalSettings,omitempty"`
}

type HealthCheck struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

type IntegrationStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

var (
	settingsMu sync.RWMutex
	settings   = &SystemSettings{
		K8sClusterURL:  "https://kubernetes.default.svc",
		DefaultRegistry: "docker.io",
		GlobalSettings:  map[string]string{},
	}
)

func GetSettings() SystemSettings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	cpy := SystemSettings{
		K8sClusterURL:  settings.K8sClusterURL,
		DefaultRegistry: settings.DefaultRegistry,
	}
	if settings.GlobalSettings != nil {
		cpy.GlobalSettings = make(map[string]string, len(settings.GlobalSettings))
		for k, v := range settings.GlobalSettings {
			cpy.GlobalSettings[k] = v
		}
	}
	return cpy
}

func UpdateSettings(updates *SystemSettings) {
	if updates == nil {
		return
	}
	settingsMu.Lock()
	defer settingsMu.Unlock()
	if updates.K8sClusterURL != "" {
		settings.K8sClusterURL = updates.K8sClusterURL
	}
	if updates.DefaultRegistry != "" {
		settings.DefaultRegistry = updates.DefaultRegistry
	}
	if updates.GlobalSettings != nil {
		if settings.GlobalSettings == nil {
			settings.GlobalSettings = make(map[string]string)
		}
		for k, v := range updates.GlobalSettings {
			settings.GlobalSettings[k] = v
		}
	}
}

func CheckHealth(db *gorm.DB, rdb *redis.Client, k8sStatus string) *HealthCheck {
	checks := make(map[string]string)
	allOK := true

	if err := database.PingPostgres(db); err != nil {
		checks["db"] = "fail"
		allOK = false
	} else {
		checks["db"] = "ok"
	}

	if rdb != nil {
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			checks["redis"] = "fail"
			allOK = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "unavailable"
	}

	if k8sStatus != "" {
		checks["k8s"] = k8sStatus
		if k8sStatus != "ok" {
			allOK = false
		}
	} else {
		checks["k8s"] = "ok"
	}

	status := "ok"
	if !allOK {
		status = "degraded"
	}
	return &HealthCheck{Status: status, Checks: checks}
}
