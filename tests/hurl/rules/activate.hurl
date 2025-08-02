# Create a new rule (inactive by default)
POST http://localhost:8888/api/rules
Content-Type: application/json

{
  "name": "testsuite.{{newUuid}}",
  "description": "A rule to test active flag",
  "source": "github",
  "event_type": "pull_request",
  "condition": "select * from events where state = 'open'",
  "evaluation_type": "LIVE_TRIGGER",
  "alert_level": "MEDIUM"
}

HTTP/1.1 201
[Asserts]
jsonpath "$.active" == false
jsonpath "$.id" exists
[Captures]
id: jsonpath "$.id"

# Patch the rule to make it active
PATCH http://localhost:8888/api/rules/{{id}}
Content-Type: application/json

{
  "active": true
}

HTTP/1.1 200
[Asserts]
jsonpath "$.active" == true
jsonpath "$.id" exists

# Get the rule and verify active is true
GET http://localhost:8888/api/rules/{{id}}

HTTP/1.1 200
[Asserts]
jsonpath "$.active" == true
jsonpath "$.id" exists

