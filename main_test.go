package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handler "json_to_qr_api/api"
)

var testWib = time.FixedZone("WIB", 7*3600)

func TestComputePayload(t *testing.T) {
	tests := []struct {
		name          string
		qrType        string
		date          time.Time
		wantID        int
		wantType      string
		wantCreatedAt string
	}{
		{
			name:          "Datang 14 Okt 2026",
			qrType:        "datang",
			date:          time.Date(2026, 10, 14, 10, 0, 0, 0, testWib),
			wantID:        4275,
			wantType:      "QR Datang",
			wantCreatedAt: "2026-10-14 03:00:00",
		},
		{
			name:          "Pulang 14 Okt 2026",
			qrType:        "pulang",
			date:          time.Date(2026, 10, 14, 10, 0, 0, 0, testWib),
			wantID:        4276,
			wantType:      "QR Pulang",
			wantCreatedAt: "2026-10-14 14:45:00",
		},
		{
			name:          "Datang 15 Okt 2026 (+1 hari)",
			qrType:        "datang",
			date:          time.Date(2026, 10, 15, 1, 0, 0, 0, testWib),
			wantID:        4276,
			wantType:      "QR Datang",
			wantCreatedAt: "2026-10-15 03:00:00",
		},
		{
			name:          "Pulang 15 Okt 2026 (+1 hari)",
			qrType:        "pulang",
			date:          time.Date(2026, 10, 15, 23, 0, 0, 0, testWib),
			wantID:        4277,
			wantType:      "QR Pulang",
			wantCreatedAt: "2026-10-15 14:45:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := handler.ComputePayload(tc.qrType, tc.date)
			if got.ID != tc.wantID {
				t.Errorf("ID got %d, want %d", got.ID, tc.wantID)
			}
			if got.LemdikID != 4 {
				t.Errorf("LemdikID got %d, want 4", got.LemdikID)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type got %s, want %s", got.Type, tc.wantType)
			}
			if got.CreatedAt != tc.wantCreatedAt {
				t.Errorf("CreatedAt got %s, want %s", got.CreatedAt, tc.wantCreatedAt)
			}
		})
	}
}

func TestHandleQRJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/qr?type=datang&date=2026-10-14", nil)
	rr := httptest.NewRecorder()

	handler.Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code got %d, want %d", rr.Code, http.StatusOK)
	}

	var resp handler.QRResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed unmarshal: %v", err)
	}

	if resp.Payload.ID != 4275 {
		t.Errorf("expected id 4275, got %d", resp.Payload.ID)
	}

	rawPNG, err := base64.StdEncoding.DecodeString(resp.QRBase64)
	if err != nil || len(rawPNG) == 0 {
		t.Errorf("invalid base64 qr payload")
	}
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code got %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
