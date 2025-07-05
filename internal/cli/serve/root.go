package serve

import (
	"database/sql"
	"github.com/turbolytics/sqlsec/internal/db"
	"github.com/turbolytics/sqlsec/internal/server"
	"github.com/turbolytics/sqlsec/internal/server/handlers"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCommand() *cobra.Command {
	var port string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the SQLSec API server",
		Run: func(cmd *cobra.Command, args []string) {
			logger, err := zap.NewProduction()
			if err != nil {
				log.Fatalf("failed to initialize zap logger: %v", err)
			}
			defer logger.Sync()

			dsn := os.Getenv("SQLSEC_DB_DSN")
			if dsn == "" {
				logger.Fatal("SQLSEC_DB_DSN environment variable not set")
			}

			dbConn, err := sql.Open("postgres", dsn)
			if err != nil {
				logger.Fatal("failed to connect to database", zap.Error(err))
			}
			defer dbConn.Close()

			queries := db.New(dbConn)
			wh := handlers.NewWebhook(queries, logger)
			nh := handlers.NewNotificationHandlers(queries)
			rh := handlers.NewRuleHandlers(queries)
			dh := handlers.NewDestinationHandlers(queries)
			router := chi.NewRouter()
			server.RegisterRoutes(router, wh, nh, rh, dh, logger)

			addr := ":" + port
			logger.Info("Starting server", zap.String("addr", addr))
			if err := http.ListenAndServe(addr, router); err != nil {
				logger.Fatal("server failed", zap.Error(err))
			}
		},
	}
	cmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	return cmd
}
