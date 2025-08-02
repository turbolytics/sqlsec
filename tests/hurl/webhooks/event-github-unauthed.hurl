# Create a webhook and extract the id and secret
POST http://localhost:8888/api/webhooks
Content-Type: application/json
{
  "name": "Test Webhook",
  "source": "github",
  "events": ["push"]
}
HTTP/1.1 201
[Asserts]
jsonpath "$.id" exists
jsonpath "$.secret" exists
[Captures]
webhook_id: jsonpath "$.id"
webhook_secret: jsonpath "$.secret"

# Prepare a fake payload and signature
# The payload is: {"action":"ping"}
# The secret is captured from the previous response
# The signature below is for demonstration and will not match the real HMAC, so the test expects 401
POST http://localhost:8888/api/events/{{webhook_id}}
Content-Type: application/json
X-Hub-Signature-256: sha256=deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef

{"action":"ping"}
HTTP/1.1 401
[Asserts]
body contains "invalid signature"

