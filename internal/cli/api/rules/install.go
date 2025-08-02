package rules

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/turbolytics/sqlsec/internal"
)

// Pre-made rules map: name -> SQL
var preMadeRules = map[string]internal.Rule{
	"github-pull-request-merged-no-reviewers": {
		Name:           "no-reviewers",
		Description:    "Detects pull requests that were merged without any reviewers",
		EvaluationType: internal.EvaluationTypeLiveTrigger,
		EventSource:    "github",
		EventType:      "pull_request",
		SQL: `
SELECT *
FROM events
WHERE
	raw_payload->>'action' == 'closed'
  	AND json_extract(raw_payload, '$.pull_request.merged') == true
  	AND json_array_length(json_extract(raw_payload, '$.pull_request.assignees')) == 0
  	AND json_array_length(json_extract(raw_payload, '$.pull_request.requested_reviewers')) == 0
  	AND json_extract(raw_payload, '$.pull_request.review_comments') == 0
  	AND json_extract(raw_payload, '$.pull_request.comments') == 0;
`,
		AlertLevel: internal.AlertLevelLow,
		Active:     true,
	},
}

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a pre-made rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, _ := cmd.Flags().GetBool("list")
			if list {
				// Print table of all preMadeRules using go-pretty, following the create.go example
				t := table.NewWriter()
				t.SetOutputMirror(cmd.OutOrStdout())
				t.AppendHeader(table.Row{"ID", "Name", "Description", "Event Source", "Event Type", "Evaluation Type", "Alert Level", "Active"})
				for id, rule := range preMadeRules {
					t.AppendRow(table.Row{
						id,
						rule.Name,
						rule.Description,
						rule.EventSource,
						rule.EventType,
						rule.EvaluationType,
						rule.AlertLevel,
						rule.Active,
					})
				}
				t.SetStyle(table.StyleDefault)
				t.Style().Options.SeparateRows = false
				t.Style().Box = table.StyleBoxDefault
				t.Style().Format.Header = text.FormatDefault
				t.Render()
				return nil
			}
			id, _ := cmd.Flags().GetString("id")
			rule, ok := preMadeRules[id]
			if !ok {
				return fmt.Errorf("Unknown rule: %s", id)
			}
			// Use doCreate and printRuleTable directly instead of cobra subcommand
			baseURL, _ := cmd.Flags().GetString("base-url")
			ruleResp, err := doCreate(
				baseURL,
				rule.Name,
				rule.Description,
				rule.EventSource,
				rule.EventType,
				rule.SQL,
				string(rule.EvaluationType),
				string(rule.AlertLevel),
				rule.Active,
			)
			if err != nil {
				return err
			}
			printRuleTable(cmd, ruleResp)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Id of the pre-made rule")
	cmd.Flags().Bool("list", false, "List all available pre-made rules")
	cmd.MarkFlagsMutuallyExclusive("id", "list")
	return cmd
}
