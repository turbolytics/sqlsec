# Monitor GitHub

## Task List

- [ ] Start sqlsec
- [ ] Create a Slack Notification Channel
    - [ ] Create a slack app with incoming webhook
    - [ ] Test the notification channel
- [ ] Create a Rule to monitor GitHub events
    - [ ] Test the rule 
    - [ ] Set the rule notification channel
- [ ] Create a Webhook to receive events from GitHub
    - [ ] Register the Webhook in GitHub


## Examples 
```
make build
export SQLSEC_API_BASE_URL=http://localhost:8888
```

### Create a Slack Notification Channel
```
./bin/sqlsec api channels create --name="github-dco-1" --type="slack" --config-webhook-url=<your-slack-webhook-url>
┌───────────┬────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Attribute │ Value                                                                                              │
├───────────┼────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ id        │ 23ca692d-cc83-43b2-9f6a-f8d9ab11d51c                                                               │
│ name      │ github-dco-1                                                                                       │
│ type      │ slack                                                                                              │
│ config    │ map[webhook_url:https://hooks.slack.com/services/                                                  │
└───────────┴────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

Test the channel:

```
./bin/sqlsec api channels test 23ca692d-cc83-43b2-9f6a-f8d9ab11d51c
Test successful: Test message sent
```

<img width="1089" height="126" alt="Screenshot 2025-07-26 at 6 14 04 AM" src="https://github.com/user-attachments/assets/9cd7f718-1926-4a67-ab52-a1fd14f6e705" />

### Create a Webhook
```
./bin/sqlsec api webhooks create --name=github-dco-1 --source=github

+------------+------------------------------------------------------------------+
| Attribute  | Value                                                            |
+------------+------------------------------------------------------------------+
| source     | github                                                           |
| created_at | 2025-07-25T10:47:17.212476Z                                      |
| events     | <nil>                                                            |
| id         | a15f65dc-35d6-41fb-a7f4-583153a08af4                             |
| tenant_id  | 00000000-0000-0000-0000-000000000000                             |
| name       | github-dco-1                                                     |
| secret     | 1ddd56351c3bae833a57f107abe14516bdb1eafb43e33e55632d0bf817fedb25 |
+------------+------------------------------------------------------------------+
```
