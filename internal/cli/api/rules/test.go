package rules

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func NewTestCmd() *cobra.Command {
	var event string

	cmd := &cobra.Command{
		Use:   "test <rule-id>",
		Short: "Test a rule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleID := args[0]
			baseURL, _ := cmd.Flags().GetString("base-url")
			url := fmt.Sprintf("%s/api/rules/%s/test", baseURL, ruleID)
			resp, err := http.Post(url, "application/json", strings.NewReader(event))
			if err != nil {
				return fmt.Errorf("failed to call test endpoint: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			fmt.Println(string(body))

			return nil
		},
	}

	cmd.Flags().StringVar(&event, "event", "{}", "Test the rule execution against the event")

	return cmd
}
