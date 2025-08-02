package destinations

/*
func doList(baseURL, ruleID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/rules/%s/destinations", baseURL, ruleID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status: %s, body: %s", resp.Status, string(respBody))
	}
	var dests []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&dests); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return dests, nil
}

func printDestinationsTable(cmd *cobra.Command, dests []map[string]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	if len(dests) == 0 {
		t.AppendRow(table.Row{"No destinations found."})
		t.Render()
		return
	}
	// Collect all keys for header
	headers := []string{}
	for k := range dests[0] {
		headers = append(headers, k)
	}
	t.AppendHeader(table.Row(headers))
	for _, dest := range dests {
		row := make(table.Row, len(headers))
		for i, h := range headers {
			row[i] = dest[h]
		}
		t.AppendRow(row)
	}
	t.Render()
}
*/
