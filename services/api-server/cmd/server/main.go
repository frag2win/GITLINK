package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/database"
	"github.com/localrepo/api-server/internal/handlers"
	"github.com/localrepo/api-server/internal/ipc"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/localrepo/api-server/internal/router"
	"github.com/localrepo/api-server/internal/service"
	"github.com/localrepo/api-server/internal/ssh"
)

func main() {
	// 1. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Load and Validate Config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 3. Open Database
	db, err := database.New(cfg.DBUrl)
	if err != nil {
		logger.Error("Failed to initialise database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Run Migrations (if applicable in dev, or skip in prod)
	if cfg.DevMode {
		logger.Info("DevMode enabled: running database migrations")
		if err := db.Conn.AutoMigrate(
			&models.User{},
			&models.SSHKey{},
			&models.Repository{},
			&models.RepositoryCollaborator{},
			&models.BranchProtection{},
			&models.PullRequest{},
			&models.AuditLog{},
			&models.SyncTask{},
			&models.PullRequestReview{},
			&models.ReviewThread{},
			&models.PullRequestReviewComment{},
			&models.Organization{},
			&models.Team{},
			&models.TeamMember{},
			&models.TeamRepositoryPermission{},
			&models.Notification{},
		); err != nil {
			logger.Error("Failed to run database migrations", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("DevMode disabled: skipping database migrations")
	}

	// 5. Wire DI Container (Repositories)
	userRepo := repository.NewUserRepository(db.Conn)
	sshRepo := repository.NewSSHRepository(db.Conn)
	repoRepo := repository.NewRepoRepository(db.Conn)
	healthRepo := repository.NewHealthRepository(db.Conn)
	contributorRepo := repository.NewContributorRepository(db.Conn)
	branchProtectionRepo := repository.NewBranchProtectionRepository(db.Conn)
	syncRepo := repository.NewSyncRepository(db.Conn)
	prReviewRepo := repository.NewPRReviewRepository(db.Conn)
	teamRepo := repository.NewTeamRepository(db.Conn)
	notifRepo := repository.NewNotificationRepository(db.Conn)

	auditRepo := repository.NewAuditRepository(db.Conn)
	txManager := repository.NewTransactionManager(db.Conn)

	// 6. Wire DI Container (Clients & EventBus)
	gitTransport, err := ipc.NewTransport(cfg.GitIPCNetwork, cfg.GitIPCAddress, 30*time.Second)
	if err != nil {
		logger.Error("Failed to initialize git IPC transport", "error", err)
		os.Exit(1)
	}
	gitClient := ipc.NewGitClient(gitTransport, 30*time.Second)
	p2pClient := ipc.NewP2PClient(cfg.P2PSocketPath, 30*time.Second)
	eventBus := service.NewEventBus()

	// 7. Wire DI Container (Services)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	sshSvc := service.NewSSHService(sshRepo)
	gitSvc := service.NewGitService(gitClient)
	auditSvc := service.NewAuditService(auditRepo)
	repoSvc := service.NewRepoService(repoRepo, gitSvc, auditSvc, txManager)
	healthSvc := service.NewHealthService(healthRepo)
	branchProtectSvc := service.NewBranchProtectionService(branchProtectionRepo)
	authzSvc := service.NewAuthorizationService(repoRepo, contributorRepo, branchProtectionRepo, teamRepo)
	prRepo := repository.NewPullRequestRepository(db.Conn)
	pullSvc := service.NewPullRequestService(prRepo, prReviewRepo, gitSvc, eventBus)
	teamSvc := service.NewTeamService(teamRepo)
	notifSvc := service.NewNotificationService(notifRepo, eventBus)
	wsHub := service.NewWebSocketHub(eventBus)
	conflictSvc := service.NewConflictService(gitSvc)

	peerSvc := service.NewPeerService(p2pClient)
	syncSvc := service.NewSyncService(syncRepo, peerSvc, logger)
	ctx, cancelSyncWorker := context.WithCancel(context.Background())
	defer cancelSyncWorker()
	syncSvc.StartWorker(ctx)

	// 8. Wire DI Container (Handlers)
	diHandlers := &router.Handlers{
		Auth:         handlers.NewAuthHandler(authSvc),
		SSH:          handlers.NewSSHHandler(sshSvc),
		Repo:         handlers.NewRepoHandler(repoSvc),
		Branch:       handlers.NewBranchHandler(gitSvc, repoSvc, branchProtectSvc, authzSvc),
		Pull:         handlers.NewPullRequestHandler(pullSvc, repoSvc),
		Commit:       handlers.NewCommitHandler(gitSvc, repoSvc),
		File:         handlers.NewFileHandler(gitSvc, repoSvc),
		Contributor:  handlers.NewContributorHandler(repoSvc, contributorRepo, userRepo),
		Health:       handlers.NewHealthHandler(healthSvc, syncRepo),
		GitHTTP:      handlers.NewGitHTTPHandler(gitSvc, authzSvc),
		Sync:         handlers.NewSyncHandler(syncSvc, peerSvc, syncRepo),
		Team:         handlers.NewTeamHandler(teamSvc),
		Notification: handlers.NewNotificationHandler(notifSvc),
		Metrics:      handlers.NewMetricsHandler(syncRepo),
		WS:           handlers.NewWSHandler(wsHub, notifSvc),
		Admin:        handlers.NewAdminHandler(syncRepo, peerSvc, auditSvc),
		Conflict:     handlers.NewConflictHandler(conflictSvc, repoSvc),
	}

	// 8. Start IPC Server for Git Hooks
	authIPC := ipc.NewAuthSocketServer(cfg.AuthSocketPath, ipc.PushAuthorizerFunc(func(ctx context.Context, userID uint, repoName, branchName string) (bool, string, error) {
		res, err := authzSvc.AuthorizePush(ctx, userID, repoName, branchName)
		if err != nil {
			return false, "", err
		}
		return res.Allowed, res.Reason, nil
	}), logger)
	go func() {
		if err := authIPC.Start(); err != nil {
			logger.Error("IPC Auth Socket server error", "error", err)
		}
	}()

	// 9. Create and Configure REST App
	app := fiber.New(fiber.Config{
		AppName:      "LocalRepo API Server",
		ServerHeader: "LocalRepo",
	})
	middleware.SetupCORS(app)
	middleware.SetupLogger(app)
	router.Setup(app, diHandlers, cfg)

	// 10. Start SSH Server
	go func() {
		if err := ssh.Start(cfg, sshSvc, repoSvc, authzSvc); err != nil {
			logger.Error("SSH server error", "error", err)
		}
	}()

	// 11. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
	}()

	// 12. Start listening REST
	addr := ":" + cfg.Port
	logger.Info("api-server listening", "address", addr)
	if err := app.Listen(addr); err != nil {
		logger.Error("REST server error", "error", err)
	}
}
