POST http://localhost:8888/api/rules/
Content-Type: application/json

{
  "name": "testsuite.Test Rule",
  "description": "A test rule for PRs",
  "source": "github",
  "event_type": "pull_request",
  "condition": "SELECT * FROM events WHERE state = 'open'",
  "evaluation_type": "LIVE_TRIGGER",
  "alert_level": "MEDIUM"
}

HTTP/1.1 201
[Asserts]
jsonpath "$.id" exists
jsonpath "$.name" == "testsuite.Test Rule"
jsonpath "$.description" == "A test rule for PRs"
jsonpath "$.event_source" == "github"
jsonpath "$.event_type" == "pull_request"
jsonpath "$.evaluation_type" == "LIVE_TRIGGER"
jsonpath "$.alert_level" == "MEDIUM"
jsonpath "$.created_at" exists

