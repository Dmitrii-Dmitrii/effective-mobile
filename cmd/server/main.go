package main

import (
	"context"
	"effective-mobile/api"
	"effective-mobile/internal/drivers"
	"effective-mobile/internal/services"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		//log.Error().Err(err).Msg(custom_errors.ErrEnvLoading.Message)
	}

	connString := os.Getenv("DB_CONNECTION_STRING")
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		fmt.Println(err)
		//log.Error().Err(err).Msg(custom_errors.ErrCreatePool.Message)
	}
	fmt.Println("connected to database")
	//log.Info().Msg("Connected to database")
	defer dbpool.Close()

	personDriver := drivers.NewPersonDriver(dbpool)
	personService := services.NewPersonService(*personDriver)
	personHandler := api.NewPersonHandler(*personService)

	router := gin.Default()

	router.POST("/persons", personHandler.CreatePerson)
	router.PUT("/persons", personHandler.UpdatePerson)
	router.GET("/persons", personHandler.GetPersons)
	router.DELETE("/persons/:id", func(c *gin.Context) {
		uuid := pgtype.UUID{}
		if err := uuid.Scan(c.Param("id")); err != nil {
			c.JSON(400, gin.H{"error": "invalid UUID"})
			return
		}
		personHandler.DeletePerson(c, uuid)
	})
	router.GET("/persons/:id", func(c *gin.Context) {
		uuid := pgtype.UUID{}
		if err := uuid.Scan(c.Param("id")); err != nil {
			c.JSON(400, gin.H{"error": "invalid UUID"})
			return
		}
		personHandler.GetPersonById(c, uuid)
	})

	server := &http.Server{
		Addr:         getServerAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		//log.Info().Msg(fmt.Sprintf("Server starting on %s", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			//log.Error().Err(err).Msg(custom_errors.ErrStartServer.Message)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	//log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		//log.Error().Err(err).Msg(custom_errors.ErrShutdownServer.Message)
	}

	//log.Info().Msg("Server exiting")
}

func getServerAddress() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
