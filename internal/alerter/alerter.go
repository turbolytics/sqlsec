package alerter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/turbolytics/sqlsec/internal/db/queries/alerts"
	"github.com/turbolytics/sqlsec/internal/db/queries/events"
	"github.com/turbolytics/sqlsec/internal/db/queries/rules"
	"github.com/turbolytics/sqlsec/internal/notify"
	"go.uber.org/zap"
	"time"
)

type Alerter struct {
	alertQueries alerts.Querier
	ruleQueries  rules.Querier
	eventQuerier events.Querier
	notifyReg    *notify.Registry
	logger       *zap.Logger
	db           *sql.DB
}

func NewAlerter(
	db *sql.DB,
	alertQ alerts.Querier,
	ruleQ rules.Querier,
	eventQ events.Querier,
	notifyReg *notify.Registry,
	logger *zap.Logger,
) *Alerter {

	return &Alerter{
		alertQueries: alertQ,
		eventQuerier: eventQ,
		ruleQueries:  ruleQ,
		notifyReg:    notifyReg,
		logger:       logger,
		db:           db,
	}
}

// Run starts the engine in daemon mode.
func (a *Alerter) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.ExecuteOnce(ctx); err != nil {
				a.logger.Error("Failed to execute alert processing", zap.Error(err))
			}
		}
	}
}

func (a *Alerter) ExecuteOnce(ctx context.Context) error {
	// 1. Fetch next alert for processing
	alertID, err := a.alertQueries.FetchNextAlertForProcessing(ctx, sql.NullString{})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // No event to process
		}
		a.logger.Error("Failed to fetch next alert for processing", zap.Error(err))
		return err
	}
	// 2. Get alert details
	alert, err := a.alertQueries.GetAlertByID(ctx, alertID)
	if err != nil {
		a.logger.Error("Failed to get alert by ID", zap.Error(err))
		return err
	}
	// 3. Get notification channels for the rule
	channels, err := a.ruleQueries.ListNotificationChannelsForRule(ctx, alert.RuleID)
	if err != nil {
		a.logger.Error("Failed to list notification channels for rule", zap.Error(err))
		return err
	}
	// 4. Send alert to each channel
	for _, ch := range channels {
		notifier, err := a.notifyReg.Get(notify.ChannelType(ch.Type))
		if err != nil {
			// A missing notifier is a fatal error, we should not continue processing
			a.logger.Fatal("No notifier for channel type", zap.String("type", ch.Type), zap.Error(err))
			panic("No notifier for channel type: " + ch.Type)
		}
		// Parse ch.Config (json.RawMessage) into map[string]any
		var cfg map[string]string
		if err := json.Unmarshal(ch.Config, &cfg); err != nil {
			a.logger.Error("Failed to parse channel config", zap.Error(err))
			continue
		}
		// TODO: Render the alert message, include the source, and the rule SQL, and the Level
		msg := notify.Message{
			Title: "Alert",
			Body:  fmt.Sprintf("Alerted from alert: %s", alert.ID.String()),
		} // TODO - Render the alert message, include the source, and the rule SQL, and the Level

		deliverErr := notifier.Send(ctx, cfg, msg)
		status := "delivered"
		if deliverErr != nil {
			status = "failed"
			a.logger.Error("Failed to deliver alert", zap.Error(deliverErr))
		}
		// Save delivery status
		err = a.alertQueries.InsertAlertDelivery(ctx, alerts.InsertAlertDeliveryParams{
			AlertID:   alert.ID,
			ChannelID: ch.ID,
			Status:    status,
		})
		if err != nil {
			a.logger.Error("Failed to insert alert delivery", zap.Error(err))
		}
	}

	// TODO - Should be transactional, only mark if all deliveries succeeded

	// 5. Update alert_processing_queue status
	err = a.alertQueries.MarkAlertProcessingDelivered(ctx, alert.ID)
	if err != nil {
		a.logger.Error("Failed to mark alert processing delivered", zap.Error(err))
	}
	// 6. Update alert to notified = true
	err = a.alertQueries.MarkAlertNotified(ctx, alert.ID)
	if err != nil {
		a.logger.Error("Failed to mark alert as notified", zap.Error(err))
	}
	return nil
}
