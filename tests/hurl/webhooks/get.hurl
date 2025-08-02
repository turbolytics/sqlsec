POST http://localhost:8888/api/webhooks
Content-Type: application/json
{
  "name": "Test Webhook",
  "source": "github",
  "events": ["event1", "event2"]
}
HTTP/1.1 201
[Captures]
webhook_id: jsonpath "$['id']"

GET http://localhost:8888/api/webhooks/{{webhook_id}}
HTTP/1.1 200