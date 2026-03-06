package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/config"
	"github.com/xjy/zcid/internal/admin"
	"github.com/xjy/zcid/internal/audit"
	"github.com/xjy/zcid/internal/crdclean"
	"github.com/xjy/zcid/internal/auth"
	"github.com/xjy/zcid/internal/deployment"
	"github.com/xjy/zcid/internal/environment"
	gitmod "github.com/xjy/zcid/internal/git"
	"github.com/xjy/zcid/internal/registry"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/internal/pipelinerun"
	"github.com/xjy/zcid/internal/project"
	"github.com/xjy/zcid/internal/rbac"
	"github.com/xjy/zcid/internal/svcdef"
	"github.com/xjy/zcid/internal/logarchive"
	"github.com/xjy/zcid/internal/notification"
	"github.com/xjy/zcid/internal/variable"
	"github.com/xjy/zcid/internal/ws"
	"github.com/xjy/zcid/pkg/cache"
	"github.com/xjy/zcid/pkg/argocd"
	"github.com/xjy/zcid/pkg/crypto"
	k8sclient "github.com/xjy/zcid/pkg/k8s"
	"github.com/xjy/zcid/pkg/tekton"
	"github.com/xjy/zcid/pkg/database"
	"github.com/xjy/zcid/pkg/logging"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
	"github.com/xjy/zcid/pkg/storage"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	logging.Init(cfg.Server.LogLevel)
	slog.Info("logger initialized", slog.String("level", logging.CurrentLevel()))

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("connected to PostgreSQL")

	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := database.RunMigrations(cfg.Database.MigrationURL(), "migrations"); err != nil {
			slog.Error("failed to run startup migrations", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("startup migrations completed")
	}

	rdb, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		slog.Error("failed to connect to Redis", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("connected to Redis")

	minioClient, err := storage.NewMinIO(&cfg.MinIO)
	if err != nil {
		slog.Error("failed to connect to MinIO", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("connected to MinIO")

	var coreClient kubernetes.Interface
	var dynClient dynamic.Interface

	if cfg.K8s.Enabled {
		slog.Info("K8s 集群集成已启用，初始化客户端...")
		k8sClients, k8sErr := k8sclient.NewClients()
		if k8sErr != nil {
			slog.Warn("K8s 客户端初始化失败，回退到 Mock 模式", slog.Any("error", k8sErr))
			cfg.K8s.Enabled = false
		} else {
			coreClient = k8sClients.CoreClient
			dynClient = k8sClients.DynamicClient
			slog.Info("K8s 客户端初始化成功")
		}
	}

	if !cfg.K8s.Enabled {
		slog.Warn("Tekton/ArgoCD: 使用 Mock 模式 (K8s 未连接)")
	}

	var crdK8s crdclean.K8sClient
	if cfg.K8s.Enabled {
		crdK8s = crdclean.NewRealK8sClient(dynClient, []string{cfg.K8s.Namespace})
	} else {
		crdK8s = &crdclean.MockK8sClient{}
	}
	cleaner := crdclean.NewCRDCleaner(crdK8s, 7)
	go cleaner.StartScheduler(context.Background(), 24*time.Hour)

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.ErrorRecovery())
	r.Use(middleware.AccessLogger())

	var aesCrypto *crypto.AESCrypto
	if cfg.Encryption.Key != "" {
		var cryptoErr error
		aesCrypto, cryptoErr = crypto.NewAESCrypto([]byte(cfg.Encryption.Key))
		if cryptoErr != nil {
			slog.Warn("AES 加密初始化失败，密钥变量功能不可用", slog.Any("error", cryptoErr))
		} else {
			slog.Info("AES-256-GCM 加密已初始化")
		}
	} else {
		slog.Warn("ZCID_ENCRYPTION_KEY 未设置，密钥变量功能不可用")
	}

	auditRepo := audit.NewRepo(db)
	auditSvc := audit.NewService(auditRepo)

	registerHealthRoutes(r, db, rdb)
	registerExampleRoutes(r)
	registerWebSocketRoutes(r, cfg.Auth.JWTSecret, cfg.K8s.Enabled, coreClient, dynClient)
	registerAuthRoutes(r, db, rdb, cfg.Auth.JWTSecret)
	registerAdminRoutes(context.Background(), r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, auditSvc)
	registerProjectRoutes(r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, minioClient, auditSvc, cfg, coreClient, dynClient)
	registerIntegrationRoutes(r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, auditSvc, cfg, coreClient, dynClient)
	registerFrontendRoutes(r)

	_ = minioClient // used during init; retained for future use

	addr := ":" + cfg.Server.Port
	slog.Info("server starting", slog.String("addr", addr), slog.String("level", logging.CurrentLevel()))
	if err := r.Run(addr); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}

func registerWebSocketRoutes(r *gin.Engine, jwtSecret string, k8sEnabled bool, coreClient kubernetes.Interface, dynClient dynamic.Interface) {
	hub := ws.NewHub()
	go hub.Run()

	var logCollector ws.LogCollector
	var k8sWatcher ws.K8sWatcher
	if k8sEnabled && coreClient != nil {
		logCollector = ws.NewRealLogCollector(coreClient)
		k8sWatcher = ws.NewRealK8sWatcher(dynClient)
		slog.Info("WebSocket: 使用真实 K8s 日志收集器和状态监听器")
	} else {
		logCollector = &ws.MockLogCollector{}
		k8sWatcher = &ws.MockK8sWatcher{}
		slog.Info("WebSocket: 使用 Mock 日志收集器和状态监听器")
	}

	logStream := ws.NewLogStreamManager(hub, logCollector, &ws.PlaceholderSecretMasker{})

	pipelineWatcher := ws.NewPipelineWatcher(hub, k8sWatcher)
	go pipelineWatcher.Start(context.Background())

	r.GET("/ws/v1/logs/:runId", ws.ServeWsLogs(hub, jwtSecret, logStream.ReplayFn(), nil))
	r.GET("/ws/v1/pipeline-status/:projectId", ws.ServeWsStatus(hub, jwtSecret, nil))
}

func registerExampleRoutes(r *gin.Engine) {
	r.GET("/example/success", func(c *gin.Context) {
		response.Success(c, gin.H{"feature": "unified-response"})
	})

	r.GET("/example/error", func(c *gin.Context) {
		err := response.NewBizError(response.CodeBadRequest, "invalid request", "demo error")
		response.HandleError(c, err)
	})

	r.GET("/example/panic", func(c *gin.Context) {
		panic(fmt.Errorf("demo panic"))
	})
}

func registerAuthRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string) {
	repo := auth.NewRepo(db, rdb)
	service := auth.NewService(repo, jwtSecret)
	handler := auth.NewHandler(service)

	v1 := r.Group("/api/v1")
	authGroup := v1.Group("/auth")
	handler.RegisterRoutes(authGroup)
}

func registerAdminRoutes(ctx context.Context, r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, auditSvc *audit.Service) {
	repo := auth.NewRepo(db, rdb)
	service := auth.NewService(repo, jwtSecret)
	handler := auth.NewHandler(service)

	enforcer, err := rbac.NewEnforcer(db)
	if err != nil {
		slog.Error("failed to initialize RBAC enforcer", slog.Any("error", err))
		os.Exit(1)
	}
	rbac.StartWatcher(ctx, enforcer, rdb)

	v1 := r.Group("/api/v1")
	adminUsers := v1.Group("/admin")
	adminUsers.Use(middleware.RequireCasbinRBAC(jwtSecret, enforcer))
	adminUsers.Use(audit.Middleware(auditSvc))
	handler.RegisterAdminUserRoutes(adminUsers)

	varRepo := variable.NewRepo(db)
	varService := variable.NewService(varRepo, aesCrypto)
	varHandler := variable.NewHandler(varService)

	adminVars := v1.Group("/admin/variables")
	adminVars.Use(middleware.RequireAdminRBAC(jwtSecret))
	adminVars.Use(audit.Middleware(auditSvc))
	varHandler.RegisterGlobalRoutes(adminVars)

	adminHandler := admin.NewAdminHandler(db, rdb)
	adminAPI := v1.Group("/admin")
	adminAPI.Use(middleware.RequireAdminRBAC(jwtSecret))
	adminAPI.Use(audit.Middleware(auditSvc))
	adminAPI.GET("/settings", adminHandler.GetSettings)
	adminAPI.PUT("/settings", adminHandler.UpdateSettings)
	adminAPI.GET("/health", adminHandler.GetHealth)
	adminAPI.GET("/integrations/status", adminHandler.GetIntegrationsStatus)

	auditHandler := audit.NewHandler(auditSvc)
	adminAPI.GET("/audit-logs", auditHandler.List)

	admin := r.Group("/admin")
	admin.Use(middleware.RequireAdminRBAC(jwtSecret))
	admin.POST("/log-level", func(c *gin.Context) {
		var req struct {
			Level string `json:"level"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.HandleError(c, response.NewBizError(response.CodeBadRequest, "invalid request", err.Error()))
			return
		}

		if err := logging.SetLevel(req.Level); err != nil {
			response.HandleError(c, response.NewBizError(response.CodeBadRequest, "invalid log level", err.Error()))
			return
		}

		response.Success(c, gin.H{"level": logging.CurrentLevel()})
	})
}

func registerProjectRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, minioClient *minio.Client, auditSvc *audit.Service, cfg *config.Config, coreClient kubernetes.Interface, dynClient dynamic.Interface) {
	projRepo := project.NewRepo(db)
	projService := project.NewService(projRepo)
	projHandler := project.NewHandler(projService)

	envRepo := environment.NewRepo(db)
	envService := environment.NewService(envRepo)
	envHandler := environment.NewHandler(envService)

	svcRepo := svcdef.NewRepo(db)
	svcService := svcdef.NewService(svcRepo)
	svcHandler := svcdef.NewHandler(svcService)

	v1 := r.Group("/api/v1")
	projects := v1.Group("/projects")
	projects.Use(middleware.JWTAuth(jwtSecret))
	projHandler.RegisterCollectionRoutes(projects)

	projectScope := projects.Group("/:id")
	projectScope.Use(middleware.RequireProjectScope(db))
	projectScope.Use(audit.Middleware(auditSvc))
	projHandler.RegisterResourceRoutes(projectScope)

	envGroup := projectScope.Group("/environments")
	envHandler.RegisterRoutes(envGroup)

	svcGroup := projectScope.Group("/services")
	svcHandler.RegisterRoutes(svcGroup)

	memberGroup := projectScope.Group("/members")
	projHandler.RegisterMemberRoutes(memberGroup)

	varRepo := variable.NewRepo(db)
	varService := variable.NewService(varRepo, aesCrypto)
	varHandler := variable.NewHandler(varService)

	varGroup := projectScope.Group("/variables")
	varHandler.RegisterProjectRoutes(varGroup)

	pipelineRepo := pipeline.NewRepo(db)
	pipelineService := pipeline.NewService(pipelineRepo)
	pipelineHandler := pipeline.NewHandler(pipelineService)

	pipelineGroup := projectScope.Group("/pipelines")
	pipelineHandler.RegisterRoutes(pipelineGroup)

	pipelineVarsGroup := pipelineGroup.Group("/:pipelineId/variables")
	varHandler.RegisterPipelineRoutes(pipelineVarsGroup)

	runRepo := pipelinerun.NewRepo(db)
	runTranslator := tekton.NewTranslator()
	var runK8s pipelinerun.K8sClient
	var runSecretInjector pipelinerun.SecretInjector
	if cfg.K8s.Enabled && dynClient != nil {
		runK8s = pipelinerun.NewRealK8sClient(dynClient)
		runSecretInjector = pipelinerun.NewRealSecretInjector(coreClient)
	} else {
		runK8s = &pipelinerun.MockK8sClient{}
		runSecretInjector = &pipelinerun.MockSecretInjector{}
	}
	runService := pipelinerun.NewService(runRepo, pipelineRepo, varService, runTranslator, runK8s, runSecretInjector)
	runHandler := pipelinerun.NewHandler(runService)
	runsGroup := pipelineGroup.Group("/:pipelineId/runs")
	runHandler.RegisterRoutes(runsGroup)

	logArchiveStorage := logarchive.NewMinIOAdapter(minioClient)
	logArchiveSvc := logarchive.NewService(logArchiveStorage, "zcid-logs")
	logArchiveHandler := logarchive.NewHandler(logArchiveSvc, nil)
	pipelineRunsGroup := projectScope.Group("/pipeline-runs/:runId")
	logArchiveHandler.RegisterRoutes(pipelineRunsGroup)

	deployRepo := deployment.NewRepo(db)
	var argoClient argocd.ArgoClient
	if cfg.ArgoCD.Enabled && cfg.ArgoCD.Server != "" {
		argoClient = argocd.NewRestClient(cfg.ArgoCD.Server, cfg.ArgoCD.Token, cfg.ArgoCD.Insecure)
		slog.Info("ArgoCD: 使用真实 REST 客户端", slog.String("server", cfg.ArgoCD.Server))
	} else {
		argoClient = &argocd.MockArgoClient{}
		slog.Info("ArgoCD: 使用 Mock 客户端")
	}
	deploySvc := deployment.NewService(deployRepo, envService, argoClient)
	deployHandler := deployment.NewHandler(deploySvc, envService)
	deployGroup := projectScope.Group("/deployments")
	deployHandler.RegisterRoutes(deployGroup)

	notifRepo := notification.NewRepo(db)
	var notifIdemCache *cache.RedisCache
	if rdb != nil {
		notifIdemCache = cache.NewRedisCache(rdb, "notification", 5*time.Minute)
	}
	notifSvc := notification.NewService(notifRepo, notifIdemCache)
	notifHandler := notification.NewHandler(notifSvc)
	notifGroup := projectScope.Group("/notification-rules")
	notifHandler.RegisterRoutes(notifGroup)

	templateGroup := v1.Group("/pipeline-templates")
	templateGroup.Use(middleware.JWTAuth(jwtSecret))
	pipelineHandler.RegisterTemplateRoutes(templateGroup)
}

func registerIntegrationRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, auditSvc *audit.Service, cfg *config.Config, coreClient kubernetes.Interface, dynClient dynamic.Interface) {
	gitRepo := gitmod.NewRepo(db)
	gitService := gitmod.NewService(gitRepo, aesCrypto)

	var gitCache *cache.RedisCache
	if rdb != nil {
		gitCache = cache.NewRedisCache(rdb, "git", 5*time.Minute)
		gitService.SetCache(gitCache)
	}

	gitHandler := gitmod.NewHandler(gitService)

	registryRepo := registry.NewRepo(db)
	registryService := registry.NewService(registryRepo, aesCrypto)
	registryHandler := registry.NewHandler(registryService)

	v1 := r.Group("/api/v1")
	integrations := v1.Group("/admin/integrations")
	integrations.Use(middleware.RequireAdminRBAC(jwtSecret))
	integrations.Use(audit.Middleware(auditSvc))
	gitHandler.RegisterRoutes(integrations)

	registries := integrations.Group("/registries")
	registryHandler.RegisterRoutes(registries)

	var idempotentCache *cache.RedisCache
	if rdb != nil {
		idempotentCache = cache.NewRedisCache(rdb, "webhook", 5*time.Minute)
	}

	// FR18: Webhook-to-pipeline matching and auto-trigger
	pipelineRepo := pipeline.NewRepo(db)
	matcher := gitmod.NewPipelineMatcher(pipelineRepo, gitRepo)
	varRepo := variable.NewRepo(db)
	varService := variable.NewService(varRepo, aesCrypto)
	runRepo := pipelinerun.NewRepo(db)
	runTranslator := tekton.NewTranslator()
	var runK8s pipelinerun.K8sClient
	var runSecretInjector pipelinerun.SecretInjector
	if cfg.K8s.Enabled && dynClient != nil {
		runK8s = pipelinerun.NewRealK8sClient(dynClient)
		runSecretInjector = pipelinerun.NewRealSecretInjector(coreClient)
	} else {
		runK8s = &pipelinerun.MockK8sClient{}
		runSecretInjector = &pipelinerun.MockSecretInjector{}
	}
	runService := pipelinerun.NewService(runRepo, pipelineRepo, varService, runTranslator, runK8s, runSecretInjector)
	webhookHandler := gitmod.NewWebhookHandler(gitService, idempotentCache, matcher, runService)
	webhooks := v1.Group("/webhooks")
	webhookHandler.RegisterRoutes(webhooks)
}

func registerHealthRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// GET /healthz — liveness probe, no dependency checks
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// GET /readyz — readiness probe, checks DB + Redis
	r.GET("/readyz", func(c *gin.Context) {
		checks := map[string]string{}
		allOK := true

		// Check DB
		if err := database.PingPostgres(db); err != nil {
			checks["db"] = "fail"
			allOK = false
		} else {
			checks["db"] = "ok"
		}

		// Check Redis
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			checks["redis"] = "fail"
			allOK = false
		} else {
			checks["redis"] = "ok"
		}

		status := "ok"
		httpStatus := http.StatusOK
		if !allOK {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status": status,
			"checks": checks,
		})
	})
}

func registerFrontendRoutes(r *gin.Engine) {
	distPath := "web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		slog.Warn("前端静态文件目录不存在，跳过前端路由", slog.String("path", distPath))
		return
	}
	r.Static("/assets", distPath+"/assets")
	r.StaticFile("/favicon.ico", distPath+"/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/ws/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}
		c.File(distPath + "/index.html")
	})
	slog.Info("前端 SPA 路由已注册", slog.String("path", distPath))
}
