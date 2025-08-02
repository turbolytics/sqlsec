package alerter

import (
	"database/sql"
	"github.com/spf13/cobra"
	alerterpkg "github.com/turbolytics/sqlsec/internal/alerter"
	"github.com/turbolytics/sqlsec/internal/db/queries/alerts"
	"github.com/turbolytics/sqlsec/internal/db/queries/events"
	"github.com/turbolytics/sqlsec/internal/db/queries/rules"
	"github.com/turbolytics/sqlsec/internal/notify"
	"github.com/turbolytics/sqlsec/internal/notify/slack"
	_ "github.com/turbolytics/sqlsec/internal/notify/slack"
	"go.uber.org/zap"
	"log"
)

func NewRunCmd(dsn *string) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the alerter service",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize DB, queries, notifier registry, logger
			ctx := cmd.Context()
			if *dsn == "" {
				log.Fatal("Postgres DSN must be set via --dsn or SQLSEC_DB_DSN env var")
			}
			db, err := sql.Open("postgres", *dsn)
			if err != nil {
				cmd.PrintErrln("Failed to connect to DB:", err)
				return
			}
			defer db.Close()

			alertQ := alerts.New(db)
			ruleQ := rules.New(db)
			eventQ := events.New(db)
			notifyReg := notify.NewRegistry()
			slack.InitializeSlack(notifyReg)

			// TODO: Register notifiers, e.g. notifyReg.Register(notify.SlackChannel, slack.New(...))
			logger, _ := zap.NewProduction()
			defer logger.Sync()

			logger.Info("Starting alerter service")
			alerter := alerterpkg.NewAlerter(
				db,
				alertQ,
				ruleQ,
				eventQ,
				notifyReg,
				logger,
			)
			if err := alerter.Run(ctx); err != nil {
				cmd.PrintErrln("Alerter error:", err)
				return
			}
		},
	}
}
