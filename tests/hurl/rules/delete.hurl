# Create a rule
POST http://localhost:8888/api/rules/
Content-Type: application/json

{
  "name": "testsuite.Delete Test Rule",
  "description": "A rule for delete testing",
  "source": "github",
  "event_type": "pull_request",
  "condition": "select * from events where state = 'open'",
  "evaluation_type": "LIVE_TRIGGER",
  "alert_level": "LOW"
}

HTTP/1.1 201
[Asserts]
jsonpath "$.id" exists
[Captures]
rule_id: jsonpath "$.id"

# Delete the rule
DELETE http://localhost:8888/api/rules/{{rule_id}}

HTTP/1.1 204

# Try to get the deleted rule
GET http://localhost:8888/api/rules/{{rule_id}}

HTTP/1.1 404
