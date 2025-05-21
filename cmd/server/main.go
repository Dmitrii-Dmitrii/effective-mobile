package main

import (
	"context"
	"effective-mobile/api"
	_ "effective-mobile/docs"
	"effective-mobile/internal/drivers"
	"effective-mobile/internal/middlerwares"
	"effective-mobile/internal/models/custom_errors"
	"effective-mobile/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func setupLogging() {
	logLevelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var logLevel zerolog.Level

	switch logLevelStr {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "panic":
		logLevel = zerolog.PanicLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	log.Info().Msgf("Log level set to %s", logLevel.String())

	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
		log.Debug().Msg("Console logging enabled for development")
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	log.Debug().Msg("Logging system initialized")
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// @title Person API
// @version 1.0
// @description API для управления данными о людях
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @schemes http https
func main() {
	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrEnvLoading.Message)
	}

	setupLogging()

	log.Info().Msg("Starting the application")

	connString := os.Getenv("DB_CONNECTION_STRING")
	if connString == "" {
		log.Fatal().Msg("Database connection string not set")
	}

	ctx := context.Background()

	log.Debug().Msg("Connecting to database")
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal().Err(err).Msg(custom_errors.ErrCreatePool.Message)
	}
	log.Info().Msg("Connected to database successfully")
	defer dbpool.Close()

	log.Debug().Msg("Initializing application components")
	personDriver := drivers.NewPersonDriver(dbpool)
	personService := services.NewPersonService(personDriver)
	personHandler := api.NewPersonHandler(personService)

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Debug().Msg("Gin set to release mode")
	} else {
		log.Debug().Msg("Gin running in debug mode")
	}

	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(middlerwares.RequestIdMiddleware())
	router.Use(middlerwares.LoggingMiddleware())

	log.Debug().Msg("Configuring API routes")

	router.POST("/persons", personHandler.CreatePerson)
	router.PUT("/persons", personHandler.UpdatePerson)
	router.GET("/persons", personHandler.GetPersons)
	router.DELETE("/persons/:id", func(c *gin.Context) {
		reqId := getRequestId(c)
		uuid := pgtype.UUID{}
		idParam := c.Param("id")

		log.Debug().
			Str("request_id", reqId).
			Str("id_param", idParam).
			Msg("Processing delete person request")

		if err := uuid.Scan(idParam); err != nil {
			log.Warn().
				Err(err).
				Str("request_id", reqId).
				Str("id_param", idParam).
				Msg("Invalid UUID format")

			c.JSON(400, gin.H{"error": "Invalid UUID format"})
			return
		}
		personHandler.DeletePerson(c, uuid)
	})
	router.GET("/persons/:id", func(c *gin.Context) {
		reqId := getRequestId(c)
		uuid := pgtype.UUID{}
		idParam := c.Param("id")

		log.Debug().
			Str("request_id", reqId).
			Str("id_param", idParam).
			Msg("Processing get person by ID request")

		if err := uuid.Scan(idParam); err != nil {
			log.Warn().
				Err(err).
				Str("request_id", reqId).
				Str("id_param", idParam).
				Msg("Invalid UUID format")

			c.JSON(400, gin.H{"error": "Invalid UUID format"})
			return
		}
		personHandler.GetPersonById(c, uuid)
	})

	router.GET("/health", func(c *gin.Context) {
		reqId := getRequestId(c)
		log.Debug().
			Str("request_id", reqId).
			Msg("Health check requested")

		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:         getServerAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("address", server.Addr).Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg(custom_errors.ErrStartServer.Message)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().
		Str("signal", sig.String()).
		Msg("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down server gracefully")
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrShutdownServer.Message)
	}

	log.Info().Msg("Server exited successfully")
}

func getServerAddress() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		log.Debug().Str("port", port).Msg("No port specified, using default")
	} else {
		log.Debug().Str("port", port).Msg("Using configured port")
	}
	return ":" + port
}

func getRequestId(c *gin.Context) string {
	reqID, exists := c.Get("RequestID")
	if !exists {
		return "unknown"
	}
	return reqID.(string)
}
