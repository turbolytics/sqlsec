POST http://localhost:8888/api/webhooks
Content-Type: application/json
{
  "name": "Test Webhook",
  "source": "github",
  "events": ["event1", "event2"]
}
HTTP/1.1 201