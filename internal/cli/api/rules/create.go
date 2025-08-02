package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

// doCreate performs the HTTP request to create a rule and returns the rule response or error.
func doCreate(baseURL, name, description, source, eventType, condition, evaluationType, alertLevel string, active bool) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"name":            name,
		"description":     description,
		"source":          source,
		"event_type":      eventType,
		"condition":       condition,
		"evaluation_type": evaluationType,
		"alert_level":     alertLevel,
		"active":          active,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(baseURL+"/api/rules", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status: %s, body: %s", resp.Status, string(respBody))
	}
	var rule map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return rule, nil
}

// printRuleTable prints a rule as a table with Attribute/Value columns.
func printRuleTable(cmd *cobra.Command, rule map[string]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.AppendHeader(table.Row{"Attribute", "Value"})
	for k, v := range rule {
		t.AppendRow(table.Row{k, v})
	}
	t.SetStyle(table.StyleDefault)
	t.Style().Options.SeparateRows = false
	t.Style().Box = table.StyleBoxDefault
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}

func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			source, _ := cmd.Flags().GetString("source")
			eventType, _ := cmd.Flags().GetString("event-type")
			condition, _ := cmd.Flags().GetString("condition")
			evaluationType, _ := cmd.Flags().GetString("evaluation-type")
			alertLevel, _ := cmd.Flags().GetString("alert-level")
			baseURL, _ := cmd.Flags().GetString("base-url")
			active, _ := cmd.Flags().GetBool("active")
			rule, err := doCreate(
				baseURL,
				name,
				description,
				source,
				eventType,
				condition,
				evaluationType,
				alertLevel,
				active,
			)
			if err != nil {
				return err
			}
			printRuleTable(cmd, rule)
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the rule")
	cmd.Flags().StringP("description", "d", "", "Description of the rule")
	cmd.Flags().StringP("source", "s", "", "Source of the rule")
	cmd.Flags().StringP("event-type", "e", "", "Event type for the rule")
	cmd.Flags().StringP("condition", "c", "", "Sql for the rule")
	cmd.Flags().StringP("evaluation-type", "t", "", "Evaluation type for the rule")
	cmd.Flags().StringP("alert-level", "a", "", "Alert level for the rule")
	cmd.Flags().BoolP("active", "A", true, "Whether the rule is active")

	return cmd
}
