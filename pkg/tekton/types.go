package tekton

// Tekton CRD types (simplified, not full K8s types)

// TypeMeta holds API version and kind
type TypeMeta struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ObjectMeta holds name, namespace, labels, etc.
type ObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PipelineRun is the Tekton PipelineRun CRD
type PipelineRun struct {
	TypeMeta   TypeMeta        `json:",inline"`
	ObjectMeta ObjectMeta      `json:"metadata"`
	Spec       PipelineRunSpec `json:"spec"`
}

// PipelineRunSpec defines the PipelineRun specification
type PipelineRunSpec struct {
	PipelineSpec PipelineSpec `json:"pipelineSpec"`
	Params       []Param      `json:"params,omitempty"`
	Workspaces   []Workspace  `json:"workspaces,omitempty"`
}

// PipelineSpec defines the embedded pipeline
type PipelineSpec struct {
	Tasks []PipelineTask `json:"tasks"`
}

// PipelineTask represents a task in the pipeline
type PipelineTask struct {
	Name     string    `json:"name"`
	TaskSpec *TaskSpec `json:"taskSpec,omitempty"`
	RunAfter []string  `json:"runAfter,omitempty"`
}

// TaskSpec defines a task's steps
type TaskSpec struct {
	Steps []Step `json:"steps"`
}

// Step represents a single step in a task
type Step struct {
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Env     []EnvVar `json:"env,omitempty"`
}

// Param represents a pipeline parameter
type Param struct {
	Name  string     `json:"name"`
	Value ParamValue `json:"value"`
}

// ParamValue holds the parameter value (string or array)
type ParamValue struct {
	Type      string   `json:"type"`
	StringVal string   `json:"stringVal,omitempty"`
	ArrayVal  []string `json:"arrayVal,omitempty"`
}

// Workspace defines a workspace declaration
type Workspace struct {
	Name string `json:"name"`
}

// WorkspaceBinding binds a workspace to a volume
type WorkspaceBinding struct {
	Name                  string    `json:"name"`
	EmptyDir              *struct{} `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *struct {
		ClaimName string `json:"claimName"`
	} `json:"persistentVolumeClaim,omitempty"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource references a secret key
type EnvVarSource struct {
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty"`
}

// SecretKeyRef references a key in a secret
type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
