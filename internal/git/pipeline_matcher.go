package git

import (
	"context"
	"strings"

	"github.com/xjy/zcid/internal/pipeline"
)

// PipelineMatcher implements PipelineMatchService for webhook-to-pipeline matching.
type PipelineMatcher struct {
	pipelineRepo   pipelineByTriggerType
	gitConnRepo    gitConnectionLister
}

type pipelineByTriggerType interface {
	ListByTriggerType(ctx context.Context, triggerType pipeline.TriggerType) ([]*pipeline.Pipeline, error)
}

type gitConnectionLister interface {
	List() ([]GitConnection, int64, error)
}

// NewPipelineMatcher creates a PipelineMatcher.
func NewPipelineMatcher(pipelineRepo pipelineByTriggerType, gitConnRepo gitConnectionLister) *PipelineMatcher {
	return &PipelineMatcher{pipelineRepo: pipelineRepo, gitConnRepo: gitConnRepo}
}

// FindMatchingPipelines returns pipelines with trigger_type=webhook where config repo
// or git connection matches the webhook repo URL. MVP: match pipelines in projects that
// have git connections matching repo URL - when connection matches, include all webhook
// pipelines; otherwise match by pipeline config repo URL.
func (m *PipelineMatcher) FindMatchingPipelines(ctx context.Context, repoURL, branch, eventType string) ([]MatchedPipeline, error) {
	_ = branch
	_ = eventType

	pipelines, err := m.pipelineRepo.ListByTriggerType(ctx, pipeline.TriggerWebhook)
	if err != nil {
		return nil, err
	}

	conns, _, err := m.gitConnRepo.List()
	if err != nil {
		return nil, err
	}

	normalizedWebhookRepo := normalizeRepoURL(repoURL)

	// Check if webhook repo URL matches any git connection (project uses that connection)
	hasMatchingConnection := false
	for _, c := range conns {
		norm := normalizeRepoURL(c.ServerURL)
		if norm != "" && strings.HasPrefix(normalizedWebhookRepo, norm) {
			hasMatchingConnection = true
			break
		}
	}

	var result []MatchedPipeline
	for _, p := range pipelines {
		pipeRepoURL := extractRepoURLFromConfig(p)
		pipeRepoNorm := normalizeRepoURL(pipeRepoURL)

		matches := hasMatchingConnection
		if pipeRepoNorm != "" {
			matches = matches || pipeRepoNorm == normalizedWebhookRepo ||
				strings.HasPrefix(normalizedWebhookRepo, pipeRepoNorm) ||
				strings.HasPrefix(pipeRepoNorm, normalizedWebhookRepo)
		}
		if matches {
			result = append(result, MatchedPipeline{PipelineID: p.ID, ProjectID: p.ProjectID})
		}
	}
	return result, nil
}

func normalizeRepoURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimSuffix(u, ".git")
	return strings.ToLower(u)
}

func extractRepoURLFromConfig(p *pipeline.Pipeline) string {
	for _, stage := range p.Config.Stages {
		for _, step := range stage.Steps {
			if step.Type == "git-clone" && step.Config != nil {
				if v, ok := step.Config["repoUrl"]; ok {
					if s, ok := v.(string); ok {
						return s
					}
				}
			}
		}
	}
	return ""
}
