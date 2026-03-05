package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

type PipelineTemplate struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Params      []ParamConfig  `json:"params"`
	Config      PipelineConfig `json:"config"`
}

type TemplateSummary struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Params      []ParamConfig `json:"params"`
}

type TemplateRegistry struct {
	templates map[string]*PipelineTemplate
}

func NewTemplateRegistry() *TemplateRegistry {
	r := &TemplateRegistry{templates: make(map[string]*PipelineTemplate)}
	r.registerBuiltinTemplates()
	return r
}

func (r *TemplateRegistry) List() []*PipelineTemplate {
	result := make([]*PipelineTemplate, 0, len(r.templates))
	for _, t := range r.templates {
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (r *TemplateRegistry) Get(id string) *PipelineTemplate {
	return r.templates[id]
}

func (r *TemplateRegistry) registerBuiltinTemplates() {
	r.templates["go-microservice"] = &PipelineTemplate{
		ID:          "go-microservice",
		Name:        "Go 微服务",
		Description: "Go 微服务标准构建流程：代码检出 → 编译测试 → 容器镜像构建 → 推送",
		Category:    "backend",
		Params: []ParamConfig{
			{Name: "repoUrl", Type: "string", Description: "Git 仓库地址", Required: true},
			{Name: "branch", Type: "string", Description: "分支名称", DefaultValue: "main", Required: true},
			{Name: "imageName", Type: "string", Description: "镜像名称（含仓库前缀）", Required: true},
			{Name: "goVersion", Type: "string", Description: "Go 版本", DefaultValue: "1.24"},
		},
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{
					ID: "checkout", Name: "代码检出",
					Steps: []StepConfig{
						{ID: "git-clone", Name: "Git Clone", Type: "git-clone", Config: map[string]any{"repoUrl": "{{.repoUrl}}", "branch": "{{.branch}}"}},
					},
				},
				{
					ID: "build", Name: "编译测试",
					Steps: []StepConfig{
						{ID: "go-test", Name: "单元测试", Type: "shell", Image: "golang:{{.goVersion}}", Command: []string{"go", "test", "./..."}},
						{ID: "go-build", Name: "编译", Type: "shell", Image: "golang:{{.goVersion}}", Command: []string{"go", "build", "-o", "app", "./cmd/server"}},
					},
				},
				{
					ID: "docker", Name: "镜像构建推送",
					Steps: []StepConfig{
						{ID: "docker-build", Name: "Docker Build & Push", Type: "kaniko", Config: map[string]any{"imageName": "{{.imageName}}", "dockerfile": "Dockerfile"}},
					},
				},
			},
		},
	}

	r.templates["java-maven"] = &PipelineTemplate{
		ID:          "java-maven",
		Name:        "Java Maven",
		Description: "Java Maven 标准构建流程：代码检出 → Maven 构建 → 容器镜像构建 → 推送",
		Category:    "backend",
		Params: []ParamConfig{
			{Name: "repoUrl", Type: "string", Description: "Git 仓库地址", Required: true},
			{Name: "branch", Type: "string", Description: "分支名称", DefaultValue: "main", Required: true},
			{Name: "imageName", Type: "string", Description: "镜像名称（含仓库前缀）", Required: true},
			{Name: "javaVersion", Type: "string", Description: "Java 版本", DefaultValue: "17"},
		},
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{
					ID: "checkout", Name: "代码检出",
					Steps: []StepConfig{
						{ID: "git-clone", Name: "Git Clone", Type: "git-clone", Config: map[string]any{"repoUrl": "{{.repoUrl}}", "branch": "{{.branch}}"}},
					},
				},
				{
					ID: "build", Name: "Maven 构建",
					Steps: []StepConfig{
						{ID: "mvn-package", Name: "Maven Package", Type: "shell", Image: "maven:3.9-eclipse-temurin-{{.javaVersion}}", Command: []string{"mvn", "clean", "package", "-DskipTests=false"}},
					},
				},
				{
					ID: "docker", Name: "镜像构建推送",
					Steps: []StepConfig{
						{ID: "docker-build", Name: "Docker Build & Push", Type: "kaniko", Config: map[string]any{"imageName": "{{.imageName}}", "dockerfile": "Dockerfile"}},
					},
				},
			},
		},
	}

	r.templates["frontend-node"] = &PipelineTemplate{
		ID:          "frontend-node",
		Name:        "前端 Node",
		Description: "前端 Node.js 标准构建流程：代码检出 → 依赖安装 → 构建 → 容器镜像构建 → 推送",
		Category:    "frontend",
		Params: []ParamConfig{
			{Name: "repoUrl", Type: "string", Description: "Git 仓库地址", Required: true},
			{Name: "branch", Type: "string", Description: "分支名称", DefaultValue: "main", Required: true},
			{Name: "imageName", Type: "string", Description: "镜像名称（含仓库前缀）", Required: true},
			{Name: "nodeVersion", Type: "string", Description: "Node.js 版本", DefaultValue: "20"},
			{Name: "buildCommand", Type: "string", Description: "构建命令", DefaultValue: "npm run build"},
		},
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{
					ID: "checkout", Name: "代码检出",
					Steps: []StepConfig{
						{ID: "git-clone", Name: "Git Clone", Type: "git-clone", Config: map[string]any{"repoUrl": "{{.repoUrl}}", "branch": "{{.branch}}"}},
					},
				},
				{
					ID: "build", Name: "依赖安装与构建",
					Steps: []StepConfig{
						{ID: "npm-install", Name: "安装依赖", Type: "shell", Image: "node:{{.nodeVersion}}", Command: []string{"npm", "ci"}},
						{ID: "npm-build", Name: "构建", Type: "shell", Image: "node:{{.nodeVersion}}", Command: []string{"sh", "-c", "{{.buildCommand}}"}},
					},
				},
				{
					ID: "docker", Name: "镜像构建推送",
					Steps: []StepConfig{
						{ID: "docker-build", Name: "Docker Build & Push", Type: "kaniko", Config: map[string]any{"imageName": "{{.imageName}}", "dockerfile": "Dockerfile"}},
					},
				},
			},
		},
	}

	r.templates["java-jar-traditional"] = &PipelineTemplate{
		ID:          "java-jar-traditional",
		Name:        "Java JAR 传统构建",
		Description: "Maven 编译打包，产物上传到 MinIO",
		Category:    "backend",
		Params: []ParamConfig{
			{Name: "repoUrl", Type: "string", Description: "Git 仓库地址", Required: true},
			{Name: "branch", Type: "string", Description: "分支名称", DefaultValue: "main", Required: true},
			{Name: "javaVersion", Type: "string", Description: "Java 版本", DefaultValue: "17"},
			{Name: "buildCommand", Type: "string", Description: "Maven 构建命令", DefaultValue: "mvn clean package -DskipTests"},
			{Name: "artifactPath", Type: "string", Description: "产物路径", DefaultValue: "target/*.jar"},
			{Name: "minioEndpoint", Type: "string", Description: "MinIO 端点", Required: true},
			{Name: "minioBucket", Type: "string", Description: "MinIO Bucket", Required: true},
			{Name: "minioSecret", Type: "string", Description: "MinIO 凭据 Secret 名称", DefaultValue: ""},
		},
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{
					ID: "checkout", Name: "代码检出",
					Steps: []StepConfig{
						{ID: "git-clone", Name: "Git Clone", Type: "git-clone", Config: map[string]any{"repoUrl": "{{.repoUrl}}", "branch": "{{.branch}}"}},
					},
				},
				{
					ID: "build", Name: "Maven 构建",
					Steps: []StepConfig{
						{ID: "mvn-build", Name: "Maven Package", Type: "shell", Image: "maven:3.9-eclipse-temurin-{{.javaVersion}}", Command: []string{"sh", "-c", "cd /workspace/source && {{.buildCommand}}"}},
					},
				},
				{
					ID: "upload", Name: "上传 MinIO",
					Steps: []StepConfig{
						{ID: "minio-upload", Name: "Upload to MinIO", Type: "shell", Image: "minio/mc:latest", Command: []string{"sh", "-c", "mc alias set minio http://{{.minioEndpoint}} $MINIO_ACCESS_KEY $MINIO_SECRET_KEY && mc cp -r {{.artifactPath}} minio/{{.minioBucket}}/"}},
					},
				},
			},
		},
	}

	r.templates["generic-docker"] = &PipelineTemplate{
		ID:          "generic-docker",
		Name:        "通用 Docker",
		Description: "通用 Docker 构建流程：代码检出 → Docker 镜像构建 → 推送",
		Category:    "general",
		Params: []ParamConfig{
			{Name: "repoUrl", Type: "string", Description: "Git 仓库地址", Required: true},
			{Name: "branch", Type: "string", Description: "分支名称", DefaultValue: "main", Required: true},
			{Name: "imageName", Type: "string", Description: "镜像名称（含仓库前缀）", Required: true},
			{Name: "dockerfile", Type: "string", Description: "Dockerfile 路径", DefaultValue: "Dockerfile"},
		},
		Config: PipelineConfig{
			SchemaVersion: "1.0",
			Stages: []StageConfig{
				{
					ID: "checkout", Name: "代码检出",
					Steps: []StepConfig{
						{ID: "git-clone", Name: "Git Clone", Type: "git-clone", Config: map[string]any{"repoUrl": "{{.repoUrl}}", "branch": "{{.branch}}"}},
					},
				},
				{
					ID: "docker", Name: "镜像构建推送",
					Steps: []StepConfig{
						{ID: "docker-build", Name: "Docker Build & Push", Type: "kaniko", Config: map[string]any{"imageName": "{{.imageName}}", "dockerfile": "{{.dockerfile}}"}},
					},
				},
			},
		},
	}
}

func validateTemplateParams(tmpl *PipelineTemplate, params map[string]string) (map[string]string, error) {
	merged := make(map[string]string)

	for _, p := range tmpl.Params {
		if p.DefaultValue != "" {
			merged[p.Name] = p.DefaultValue
		}
	}
	for k, v := range params {
		merged[k] = v
	}

	var missing []string
	for _, p := range tmpl.Params {
		if p.Required {
			val, ok := merged[p.Name]
			if !ok || strings.TrimSpace(val) == "" {
				missing = append(missing, p.Name)
			}
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("缺少必填参数: %s", strings.Join(missing, ", "))
	}

	for k, v := range merged {
		if strings.ContainsAny(v, "\"\\\n\r\t") {
			return nil, fmt.Errorf("参数 '%s' 包含非法字符", k)
		}
	}

	return merged, nil
}

func applyTemplateParams(config PipelineConfig, params map[string]string) PipelineConfig {
	replaceInString := func(s string) string {
		for key, value := range params {
			s = strings.ReplaceAll(s, "{{."+key+"}}", value)
		}
		return s
	}

	replaceInSlice := func(ss []string) []string {
		result := make([]string, len(ss))
		for i, s := range ss {
			result[i] = replaceInString(s)
		}
		return result
	}

	replaceInMap := func(m map[string]string) map[string]string {
		if m == nil {
			return nil
		}
		result := make(map[string]string, len(m))
		for k, v := range m {
			result[k] = replaceInString(v)
		}
		return result
	}

	replaceInAnyMap := func(m map[string]any) map[string]any {
		if m == nil {
			return nil
		}
		result := make(map[string]any, len(m))
		for k, v := range m {
			if s, ok := v.(string); ok {
				result[k] = replaceInString(s)
			} else {
				result[k] = v
			}
		}
		return result
	}

	var newStages []StageConfig
	for _, stage := range config.Stages {
		newStage := StageConfig{
			ID:   stage.ID,
			Name: replaceInString(stage.Name),
		}
		for _, step := range stage.Steps {
			newStep := StepConfig{
				ID:      step.ID,
				Name:    replaceInString(step.Name),
				Type:    step.Type,
				Image:   replaceInString(step.Image),
				Command: replaceInSlice(step.Command),
				Args:    replaceInSlice(step.Args),
				Env:     replaceInMap(step.Env),
				Config:  replaceInAnyMap(step.Config),
			}
			newStage.Steps = append(newStage.Steps, newStep)
		}
		newStages = append(newStages, newStage)
	}

	return PipelineConfig{
		SchemaVersion: config.SchemaVersion,
		Stages:        newStages,
		Params:        config.Params,
		Metadata:      config.Metadata,
	}
}

func ToTemplateSummary(t *PipelineTemplate) TemplateSummary {
	return TemplateSummary{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Category:    t.Category,
		Params:      t.Params,
	}
}
