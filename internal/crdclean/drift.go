package crdclean

import (
	"context"
)

type ArgoClient interface {
	GetAppStatus(ctx context.Context, appName string) (sync, health string, err error)
}

type DriftReport struct {
	Drifted bool   `json:"drifted"`
	Details string `json:"details"`
}

type DriftDetector struct {
	ArgoClient ArgoClient
}

func NewDriftDetector(argo ArgoClient) *DriftDetector {
	return &DriftDetector{ArgoClient: argo}
}

func (d *DriftDetector) CheckDrift(ctx context.Context, appName string) (*DriftReport, error) {
	if d.ArgoClient == nil {
		return &DriftReport{Drifted: false, Details: "ArgoCD client not configured"}, nil
	}
	sync, health, err := d.ArgoClient.GetAppStatus(ctx, appName)
	if err != nil {
		return nil, err
	}
	drifted := sync != "Synced" || (health != "Healthy" && health != "Progressing")
	details := "Sync: " + sync + ", Health: " + health
	if drifted {
		details = "Application has drifted - " + details
	}
	return &DriftReport{Drifted: drifted, Details: details}, nil
}
