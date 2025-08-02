# Monitor GitHub

## Task List

- [ ] Start sqlsec
- [ ] Create a Slack Notification Channel
  - [ ] Create a slack app with incoming webhook
  - [ ] Test the notification channel
- [ ] Create a Webhook to receive events from GitHub
  - [ ] Register the Webhook in GitHub
- [ ] Create a Rule to monitor GitHub events
  - [ ] Test the rule 
  - [ ] Set the rule notification channel
- [ ] Trigger and event and verify the notification

## Examples 
```
make build
export SQLSEC_API_BASE_URL=http://localhost:8888

make build && \
CPATH=/opt/homebrew/Cellar/tomlplusplus/3.4.0/include \
SQLSEC_DB_DSN=postgres://sqlsec:sqlsec@localhost:5432/sqlsec?sslmode=disable \
 ./bin/sqlsec serve -p 8888
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

Create a Webhook to receive events from GitHub:
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

Expose your local server to the internet using a tool like `ngrok`:

```
ngrok http http://localhost:8888
```

Register the Webhook in GitHub:

<img width="1512" height="785" alt="Screenshot 2025-07-26 at 1 32 55 PM" src="https://github.com/user-attachments/assets/d79cb913-0bdb-433e-bb0f-3505fff9d4c2" />

Perform a test action in github to verify the Webhook

### Create a Rule to monitor Unreviewed Pull Requests

Install a pre-defined rule to monitor unreviewed pull requests:

```
./bin/sqlsec api rules install --id=github-pull-request-merged-no-reviewers
+-----------------+----------------------------------------------------------------------------------+
| Attribute       | Value                                                                            |
+-----------------+----------------------------------------------------------------------------------+
| description     | Detects pull requests that were merged without any reviewers                     |
| event_source    | github                                                                           |
| event_type      | github.pull_request                                                              |
| sql             |                                                                                  |
|                 | SELECT *                                                                         |
|                 | FROM events                                                                      |
|                 | WHERE                                                                            |
|                 |   raw_payload->>'action' = 'closed'                                              |
|                 |   AND (raw_payload->'pull_request'->>'merged')::boolean = true                   |
|                 |   AND jsonb_array_length(raw_payload->'pull_request'->'assignees') = 0           |
|                 |   AND jsonb_array_length(raw_payload->'pull_request'->'requested_reviewers') = 0 |
|                 |   AND (raw_payload->'pull_request'->>'comments')::int = 0;                       |
|                 |                                                                                  |
| name            | no-reviewers                                                                     |
| evaluation_type | LIVE_TRIGGER                                                                     |
| created_at      | 2025-07-27T10:54:35.859512Z                                                      |
| alert_level     | LOW                                                                              |
| active          | false                                                                            |
| id              | e93c8c06-a8ad-40b0-952d-9b8ced7e0095                                             |
+-----------------+----------------------------------------------------------------------------------+
```

Test the rule:

```
./bin/sqlsec api rules test e93c8c06-a8ad-40b0-952d-9b8ced7e0095
```

Set the rule notification channel:

``` 
./bin/sqlsec api rules destinations add <rule-id> <notification-channel-id> 
```
