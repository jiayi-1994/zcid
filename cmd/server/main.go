package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	osSignal "os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/config"
	"github.com/xjy/zcid/internal/admin"
	"github.com/xjy/zcid/internal/analytics"
	"github.com/xjy/zcid/internal/audit"
	"github.com/xjy/zcid/internal/auth"
	"github.com/xjy/zcid/internal/crdclean"
	"github.com/xjy/zcid/internal/deployment"
	"github.com/xjy/zcid/internal/environment"
	gitmod "github.com/xjy/zcid/internal/git"
	"github.com/xjy/zcid/internal/logarchive"
	"github.com/xjy/zcid/internal/notification"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/internal/pipelinerun"
	"github.com/xjy/zcid/internal/project"
	"github.com/xjy/zcid/internal/rbac"
	"github.com/xjy/zcid/internal/registry"
	healthsignal "github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/internal/stepexec"
	"github.com/xjy/zcid/internal/svcdef"
	"github.com/xjy/zcid/internal/variable"
	"github.com/xjy/zcid/internal/ws"
	"github.com/xjy/zcid/pkg/argocd"
	"github.com/xjy/zcid/pkg/cache"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/database"
	k8sclient "github.com/xjy/zcid/pkg/k8s"
	"github.com/xjy/zcid/pkg/logging"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
	"github.com/xjy/zcid/pkg/storage"
	"github.com/xjy/zcid/pkg/tekton"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func main() {
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

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
	go cleaner.StartScheduler(appCtx, 24*time.Hour)

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.ErrorRecovery())
	r.Use(middleware.PrometheusMetrics())
	r.Use(middleware.AccessLogger())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
	var stepRepo stepexec.Repository
	if db.Migrator().HasTable("step_executions") {
		stepRepo = stepexec.NewRepo(db)
		stepRetention := stepexec.NewRetentionWorker(stepRepo, cfg.StepExecutions.RetentionDays)
		go stepRetention.Run(appCtx)
	} else {
		slog.Warn("step execution persistence disabled; migration 000019 has not created step_executions")
	}
	var stepRecorder *stepexec.Recorder
	if cfg.K8s.Enabled && dynClient != nil && stepRepo != nil {
		stepRecorder = stepexec.NewRecorder(stepRepo, 1000)
		go stepRecorder.Run(appCtx)
		slog.Info("step execution recorder enabled")
	} else if cfg.K8s.Enabled && dynClient != nil {
		slog.Info("step execution recorder disabled until step_executions migration is applied")
	} else {
		slog.Info("step execution recorder disabled in mock mode")
	}

	registerHealthRoutes(r, db, rdb)
	registerExampleRoutes(r)
	pipelineWatcher := registerWebSocketRoutes(appCtx, r, cfg.Auth.JWTSecret, cfg.K8s.Enabled, coreClient, dynClient, stepRecorder)
	authRepo := auth.NewRepo(db, rdb)
	authService := auth.NewService(authRepo, cfg.Auth.JWTSecret)
	authService.SetAuditRecorder(auditSvc)
	tokenService := auth.NewTokenService(authRepo, auditSvc)
	registerAuthRoutes(r, rdb, authService)
	signalSvc := healthsignal.NewService(healthsignal.NewRepo(db))
	registerAdminRoutes(appCtx, r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, auditSvc, authService, tokenService, signalSvc)
	registerProjectRoutes(appCtx, r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, minioClient, auditSvc, cfg, coreClient, dynClient, stepRepo, pipelineWatcher, tokenService, signalSvc)
	registerIntegrationRoutes(r, db, rdb, cfg.Auth.JWTSecret, aesCrypto, auditSvc, cfg, coreClient, dynClient, stepRepo, tokenService, signalSvc)
	registerFrontendRoutes(r)

	_ = minioClient // used during init; retained for future use

	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", slog.String("addr", addr), slog.String("level", logging.CurrentLevel()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	osSignal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))
	appCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("server exited gracefully")
}

func registerWebSocketRoutes(ctx context.Context, r *gin.Engine, jwtSecret string, k8sEnabled bool, coreClient kubernetes.Interface, dynClient dynamic.Interface, stepRecorder *stepexec.Recorder) *ws.PipelineWatcher {
	hub := ws.NewHub()
	go hub.Run()

	var logCollector ws.LogCollector
	var k8sWatcher ws.K8sWatcher
	if k8sEnabled && coreClient != nil {
		logCollector = ws.NewRealLogCollector(coreClient)
		k8sWatcher = ws.NewRealK8sWatcher(dynClient, stepRecorder)
		slog.Info("WebSocket: 使用真实 K8s 日志收集器和状态监听器")
	} else {
		logCollector = &ws.MockLogCollector{}
		k8sWatcher = &ws.MockK8sWatcher{}
		slog.Info("WebSocket: 使用 Mock 日志收集器和状态监听器")
	}

	logStream := ws.NewLogStreamManager(hub, logCollector, &ws.PlaceholderSecretMasker{})

	pipelineWatcher := ws.NewPipelineWatcher(hub, k8sWatcher)
	go pipelineWatcher.Start(ctx)

	r.GET("/ws/v1/logs/:runId", ws.ServeWsLogs(hub, jwtSecret, logStream.ReplayFn(), nil))
	r.GET("/ws/v1/pipeline-status/:projectId", ws.ServeWsStatus(hub, jwtSecret, nil))
	return pipelineWatcher
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

func registerAuthRoutes(r *gin.Engine, rdb *redis.Client, service *auth.Service) {
	handler := auth.NewHandler(service)
	if token, generated, err := service.EnsureBootstrapToken(context.Background()); err != nil {
		slog.Error("failed to ensure bootstrap token", slog.Any("error", err))
	} else if generated {
		slog.Warn("zcid first-admin bootstrap token generated", slog.String("token", token), slog.Duration("ttl", auth.BootstrapTokenTTL))
	}

	authLimiter := middleware.NewRateLimiter(rdb, 20, time.Minute)

	v1 := r.Group("/api/v1")
	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.RateLimit(authLimiter))
	handler.RegisterRoutes(authGroup)
}

func registerAdminRoutes(ctx context.Context, r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, auditSvc *audit.Service, service *auth.Service, tokenService *auth.TokenService, signalSvc *healthsignal.Service) {
	handler := auth.NewHandler(service)
	tokenHandler := auth.NewTokenHandler(tokenService)

	enforcer, err := rbac.NewEnforcer(db)
	if err != nil {
		slog.Error("failed to initialize RBAC enforcer", slog.Any("error", err))
		os.Exit(1)
	}
	rbac.StartWatcher(ctx, enforcer, rdb)

	v1 := r.Group("/api/v1")
	adminUsers := v1.Group("/admin")
	adminUsers.Use(middleware.AdminJWTOrTokenReadAuth(jwtSecret, tokenService, auth.ScopeAdminRead))
	adminUsers.Use(audit.Middleware(auditSvc))
	handler.RegisterAdminUserRoutes(adminUsers)
	tokenHandler.RegisterRoutes(adminUsers)

	varRepo := variable.NewRepo(db)
	varService := variable.NewService(varRepo, aesCrypto)
	varHandler := variable.NewHandler(varService)

	adminVars := v1.Group("/admin/variables")
	adminVars.Use(middleware.AdminJWTOrTokenReadAuth(jwtSecret, tokenService, auth.ScopeAdminRead))
	adminVars.Use(audit.Middleware(auditSvc))
	varHandler.RegisterGlobalRoutes(adminVars)

	adminHandler := admin.NewAdminHandler(db, rdb)
	adminHandler.SetSignalService(signalSvc)
	adminAPI := v1.Group("/admin")
	adminAPI.Use(middleware.AdminJWTOrTokenReadAuth(jwtSecret, tokenService, auth.ScopeAdminRead))
	adminAPI.Use(audit.Middleware(auditSvc))
	adminAPI.GET("/settings", adminHandler.GetSettings)
	adminAPI.PUT("/settings", adminHandler.UpdateSettings)
	adminAPI.GET("/health", adminHandler.GetHealth)
	adminAPI.GET("/integrations/status", adminHandler.GetIntegrationsStatus)

	auditHandler := audit.NewHandler(auditSvc)
	adminAPI.GET("/audit-logs", auditHandler.List)

	admin := r.Group("/admin")
	admin.Use(middleware.AdminJWTOrTokenReadAuth(jwtSecret, tokenService, auth.ScopeAdminRead))
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

func registerProjectRoutes(ctx context.Context, r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, minioClient *minio.Client, auditSvc *audit.Service, cfg *config.Config, coreClient kubernetes.Interface, dynClient dynamic.Interface, stepRepo stepexec.Repository, pipelineWatcher *ws.PipelineWatcher, tokenService *auth.TokenService, signalSvc *healthsignal.Service) {
	projRepo := project.NewRepo(db)
	projService := project.NewService(projRepo, pipelineWatcher)
	registerExistingProjectNamespaces(ctx, projRepo, pipelineWatcher)
	projHandler := project.NewHandler(projService)

	envRepo := environment.NewRepo(db)
	envService := environment.NewService(envRepo)
	envService.SetSignalService(signalSvc)
	envHandler := environment.NewHandler(envService)

	svcRepo := svcdef.NewRepo(db)
	svcService := svcdef.NewService(svcRepo)
	svcHandler := svcdef.NewHandler(svcService)

	v1 := r.Group("/api/v1")
	projects := v1.Group("/projects")
	projects.Use(middleware.JWTOrMappedTokenAuth(jwtSecret, tokenService, projectTokenScopeForRoute))
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

	notifRepo := notification.NewRepo(db)
	var notifIdemCache *cache.RedisCache
	if rdb != nil {
		notifIdemCache = cache.NewRedisCache(rdb, "notification", 5*time.Minute)
	}
	notifSvc := notification.NewService(notifRepo, notifIdemCache)
	notifSvc.SetCrypto(aesCrypto)
	notifSvc.SetSlackBaseURL(cfg.Notification.SlackBaseURL)

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
	runService := pipelinerun.NewService(runRepo, pipelineRepo, varService, runTranslator, runK8s, runSecretInjector, stepRepo)
	runService.SetSignalService(signalSvc)
	runService.SetNotificationService(notifSvc)
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
	deploySvc.SetSignalService(signalSvc)
	deploySvc.SetNotificationService(notifSvc)
	deployHandler := deployment.NewHandler(deploySvc, envService)
	deployGroup := projectScope.Group("/deployments")
	deployHandler.RegisterRoutes(deployGroup)

	notifHandler := notification.NewHandler(notifSvc)
	notifGroup := projectScope.Group("/notification-rules")
	notifHandler.RegisterRoutes(notifGroup)

	analyticsSvc := analytics.NewService(analytics.NewRepo(db))
	analyticsHandler := analytics.NewHandler(analyticsSvc)
	analyticsGroup := projectScope.Group("/analytics")
	analyticsHandler.RegisterRoutes(analyticsGroup)

	templateGroup := v1.Group("/pipeline-templates")
	templateGroup.Use(middleware.JWTAuth(jwtSecret))
	pipelineHandler.RegisterTemplateRoutes(templateGroup)
}

func registerExistingProjectNamespaces(ctx context.Context, repo *project.Repo, pipelineWatcher *ws.PipelineWatcher) {
	if pipelineWatcher == nil {
		return
	}
	projects, total, err := repo.List(ctx, 1, 10000)
	if err != nil {
		slog.Warn("failed to list projects for PipelineWatcher namespace registration", slog.Any("error", err))
		return
	}
	for _, p := range projects {
		pipelineWatcher.RegisterNamespaceProject(project.DefaultRunNamespace, p.ID)
	}
	slog.Info("registered project namespaces for PipelineWatcher", slog.Int64("projects", total), slog.String("namespace", project.DefaultRunNamespace))
	_ = ctx
}

func projectTokenScopeForRoute(c *gin.Context) (string, bool) {
	path := c.FullPath()
	method := c.Request.Method
	switch {
	case method == http.MethodPost && strings.HasSuffix(path, "/pipelines/:pipelineId/runs"):
		return auth.ScopePipelinesTrigger, true
	case method == http.MethodPost && strings.HasSuffix(path, "/pipelines/from-template"):
		return auth.ScopePipelinesTrigger, true
	case method == http.MethodGet && strings.Contains(path, "/pipelines/:pipelineId/runs"):
		return auth.ScopePipelinesRead, true
	case method == http.MethodGet && strings.Contains(path, "/pipeline-runs/:runId"):
		return auth.ScopePipelinesRead, true
	case method == http.MethodPost && strings.HasSuffix(path, "/deployments"):
		return auth.ScopeDeploymentsWrite, true
	case (method == http.MethodPut || method == http.MethodDelete) && strings.Contains(path, "/deployments"):
		return auth.ScopeDeploymentsWrite, true
	case method == http.MethodGet && strings.Contains(path, "/deployments"):
		return auth.ScopeDeploymentsRead, true
	case method == http.MethodGet && strings.Contains(path, "/variables"):
		return auth.ScopeVariablesRead, true
	case method == http.MethodGet && strings.Contains(path, "/notification-rules"):
		return auth.ScopeNotificationsRead, true
	case method == http.MethodGet && strings.Contains(path, "/analytics"):
		return auth.ScopePipelinesRead, true
	default:
		return "", false
	}
}

func registerIntegrationRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, jwtSecret string, aesCrypto *crypto.AESCrypto, auditSvc *audit.Service, cfg *config.Config, coreClient kubernetes.Interface, dynClient dynamic.Interface, stepRepo stepexec.Repository, tokenService *auth.TokenService, signalSvc *healthsignal.Service) {
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
	registryService.SetSignalService(signalSvc)
	registryHandler := registry.NewHandler(registryService)

	v1 := r.Group("/api/v1")
	integrations := v1.Group("/admin/integrations")
	integrations.Use(middleware.AdminJWTOrTokenReadAuth(jwtSecret, tokenService, auth.ScopeAdminRead))
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
	notifRepo := notification.NewRepo(db)
	var notifIdemCache *cache.RedisCache
	if rdb != nil {
		notifIdemCache = cache.NewRedisCache(rdb, "notification", 5*time.Minute)
	}
	notifSvc := notification.NewService(notifRepo, notifIdemCache)
	notifSvc.SetCrypto(aesCrypto)
	notifSvc.SetSlackBaseURL(cfg.Notification.SlackBaseURL)
	var runK8s pipelinerun.K8sClient
	var runSecretInjector pipelinerun.SecretInjector
	if cfg.K8s.Enabled && dynClient != nil {
		runK8s = pipelinerun.NewRealK8sClient(dynClient)
		runSecretInjector = pipelinerun.NewRealSecretInjector(coreClient)
	} else {
		runK8s = &pipelinerun.MockK8sClient{}
		runSecretInjector = &pipelinerun.MockSecretInjector{}
	}
	runService := pipelinerun.NewService(runRepo, pipelineRepo, varService, runTranslator, runK8s, runSecretInjector, stepRepo)
	runService.SetSignalService(signalSvc)
	runService.SetNotificationService(notifSvc)
	webhookLimiter := middleware.NewRateLimiter(rdb, 60, time.Minute)
	webhookHandler := gitmod.NewWebhookHandler(gitService, idempotentCache, matcher, runService)
	webhooks := v1.Group("/webhooks")
	webhooks.Use(middleware.RateLimit(webhookLimiter))
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
