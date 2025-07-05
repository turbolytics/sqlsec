# Create a rule
POST http://localhost:8888/api/rules/
Content-Type: application/json

{
  "name": "List Test Rule",
  "description": "A rule for list testing",
  "source": "github",
  "event_type": "pull_request",
  "condition": "state = 'open'",
  "evaluation_type": "LIVE_TRIGGER",
  "alert_level": "LOW"
}

HTTP/1.1 201
[Asserts]
jsonpath "$.id" exists
jsonpath "$.name" == "List Test Rule"

# List rules
GET http://localhost:8888/api/rules/

HTTP/1.1 200
[Asserts]
jsonpath "$" count > 0
jsonpath "$[0].id" exists
