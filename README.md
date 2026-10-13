# JSON to QR API

Dynamic zero-cron QR Code generator service written in Go.

## Features
- Dynamic ID & Date calculation in WIB (UTC+7) relative to base date `2026-10-14` (ID 4275 for Datang, 4276 for Pulang).
- Zero-cron architecture: safe for serverless/cold-start environments (Render Free Tier).
- Outputs JSON with Base64 QR code and Data URI, or raw PNG image.

## Endpoints

### 1. `GET /qr`
Query parameters:
- `type`: `datang` (default) or `pulang`
- `date`: `YYYY-MM-DD` (optional, defaults to current date in WIB)
- `format`: `image` (returns PNG) or default JSON

**Example JSON Response:**
```json
{
  "payload": {
    "id": 4275,
    "lemdikId": 4,
    "createdAt": "2026-10-14 03:00:00",
    "type": "QR Datang"
  },
  "raw_json": "{\"id\":4275,\"lemdikId\":4,\"createdAt\":\"2026-10-14 03:00:00\",\"type\":\"QR Datang\"}",
  "qr_base64": "...",
  "qr_data_uri": "data:image/png;base64,..."
}
```

### 2. `GET /healthz`
Health check endpoint (`{"status":"ok"}`).
