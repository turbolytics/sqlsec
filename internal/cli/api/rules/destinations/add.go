package destinations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func doAdd(baseURL, ruleID, channelID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"rule_id":    ruleID,
		"channel_id": channelID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/api/rules/%s/destinations", baseURL, ruleID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status: %s, body: %s", resp.Status, string(respBody))
	}
	var dest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&dest); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return dest, nil
}

func printDestinationTable(cmd *cobra.Command, dest map[string]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.AppendHeader(table.Row{"Attribute", "Value"})
	for k, v := range dest {
		t.AppendRow(table.Row{k, v})
	}
	t.Render()
}

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a destination to a rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires rule_id and channel_id arguments")
			}
			baseURL, _ := cmd.Flags().GetString("base-url")
			ruleID := args[0]
			channelID := args[1]
			dest, err := doAdd(baseURL, ruleID, channelID)
			if err != nil {
				return err
			}
			printDestinationTable(cmd, dest)
			return nil
		},
	}
	return cmd
}
