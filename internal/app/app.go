package app

import (
	"context"
	"feedback-bot/config"
	"feedback-bot/internal/handler"
	"feedback-bot/internal/repo"
	"feedback-bot/internal/service"
	"feedback-bot/pkg/postgres"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Run(configPath string) {
	// Configuration
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("app - Run - config error: %v", err)
	}

	// miniapp for health checking
	healthcheck(cfg.Port)

	// Logger
	log := logrus.New()
	logrusLevel, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrusLevel)
	}
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)
	log.Info("Init application...")

	// DB
	log.Info("Init postgres...")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.MaxPoolSize))
	if err != nil {
		log.Fatalf("app - Run - postgres error: %v", err)
	}
	defer pg.Close()

	// Repositories
	log.Info("Init repositories...")
	repositories := repo.NewRepositories(pg)

	// Services
	log.Info("Init services...")
	services := service.NewServices(log, &service.ServicesDependencies{
		Repos: repositories,
		// ...
	})

	// TG bot
	log.Info("Init Telegarm bot...")
	bot := runBot(cfg.TG, services)
	log.Infof("Authorized on account %s", bot.Self.UserName)

	// Server healthy
	log.Info("Init HTTP server...")
	srv := runServer(cfg.HTTP.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("STOP signal received")

	// TG bot stopping
	log.Info("Bot stopping...")
	bot.StopReceivingUpdates()
	log.Info("Bot stopped")

	// Wait 5 sec
	log.Info("Waitting...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Server stopping
	log.Info("Server stopping...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Info("Server stopped")
}

/**
 * Telegram bot
 */
func runBot(cfgTg config.TG, services *service.Services) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(cfgTg.Token)
	if err != nil {
		services.Log.Fatalf("app - Run - init TG bot error: %v", err)
	}

	updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: 60,
	})
	messageHandler := handler.NewHandler(cfgTg, bot, services)

	go func() {
		for update := range updates {
			messageHandler.Handle(update)
		}
	}()

	return bot
}

/**
 * Health monitoring and metrics
 */
func runServer(port string) *http.Server {
	type MetricsResponse struct {
		Goroutines    int    `json:"goroutines"`
		MemoryUsageMB uint64 `json:"memory_usage_mb"`
		CPUCores      int    `json:"cpu_cores"`
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/metrics", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		c.JSON(http.StatusOK, MetricsResponse{
			Goroutines:    runtime.NumGoroutine(),
			MemoryUsageMB: m.Alloc / 1024 / 1024,
			CPUCores:      runtime.NumCPU(),
		})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("app - Run - listen error: %s\n", err)
		}
	}()

	return srv
}

func healthcheck(port string) {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		resp, err := http.Get("http://localhost:" + port + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
