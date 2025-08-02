POST http://localhost:8888/api/webhooks
Content-Type: application/json
{
  "name": "Test Webhook Invalid Source",
  "source": "notasource",
  "events": ["event1"]
}
HTTP/1.1 400
[Asserts]
header "Content-Type" == "text/plain; charset=utf-8"
body contains "unsupported source"
