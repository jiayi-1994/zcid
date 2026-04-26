package analytics

type Summary struct {
	TotalRuns        int64   `json:"totalRuns"`
	SuccessRate      float64 `json:"successRate"`
	MedianDurationMS int64   `json:"medianDurationMs"`
	P95DurationMS    int64   `json:"p95DurationMs"`
}

type DailyStat struct {
	Date             string  `json:"date"`
	Total            int64   `json:"total"`
	Succeeded        int64   `json:"succeeded"`
	Failed           int64   `json:"failed"`
	MedianDurationMS int64   `json:"medianDurationMs"`
	SuccessRate      float64 `json:"successRate"`
}

type TopFailingStep struct {
	StepName     string  `json:"stepName"`
	TaskRunName  string  `json:"taskRunName"`
	FailureCount int64   `json:"failureCount"`
	TotalCount   int64   `json:"totalCount"`
	FailureRate  float64 `json:"failureRate"`
}

type TopPipeline struct {
	PipelineID   string  `json:"pipelineId"`
	PipelineName string  `json:"pipelineName"`
	RunCount     int64   `json:"runCount"`
	SuccessRate  float64 `json:"successRate"`
}

type Response struct {
	Range           string           `json:"range"`
	Summary         Summary          `json:"summary"`
	DailyStats      []DailyStat      `json:"dailyStats"`
	TopFailingSteps []TopFailingStep `json:"topFailingSteps"`
	TopPipelines    []TopPipeline    `json:"topPipelines"`
}
