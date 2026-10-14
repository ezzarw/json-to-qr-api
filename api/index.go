package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
)

var wib = time.FixedZone("WIB", 7*3600)
var baseDate = time.Date(2026, 10, 14, 0, 0, 0, 0, wib)

const (
	baseIDDatang = 4275
	baseIDPulang = 4276
	lemdikID     = 4
)

type QRPayload struct {
	ID        int    `json:"id"`
	LemdikID  int    `json:"lemdikId"`
	CreatedAt string `json:"createdAt"`
	Type      string `json:"type"`
}

type QRResponse struct {
	Payload   QRPayload `json:"payload"`
	RawJSON   string    `json:"raw_json"`
	QRBase64  string    `json:"qr_base64"`
	QRDataURI string    `json:"qr_data_uri"`
}

func ComputePayload(qrType string, refTime time.Time) QRPayload {
	// ponytail: kalkulasi selisih hari deterministik terhadap 2026-10-14 WIB
	tWIB := refTime.In(wib)
	todayMidnight := time.Date(tWIB.Year(), tWIB.Month(), tWIB.Day(), 0, 0, 0, 0, wib)
	dayDiff := int(todayMidnight.Sub(baseDate).Hours() / 24)

	isPulang := strings.EqualFold(qrType, "pulang")

	id := baseIDDatang + dayDiff
	typeName := "QR Datang"
	timeStr := "03:00:00"

	if isPulang {
		id = baseIDPulang + dayDiff
		typeName = "QR Pulang"
		timeStr = "14:45:00"
	}

	createdAt := fmt.Sprintf("%s %s", todayMidnight.Format("2006-01-02"), timeStr)

	return QRPayload{
		ID:        id,
		LemdikID:  lemdikID,
		CreatedAt: createdAt,
		Type:      typeName,
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api/health" || path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	q := r.URL.Query()
	qrType := q.Get("type")
	if qrType == "" {
		qrType = "datang"
	}

	targetTime := time.Now().In(wib)
	if customDate := q.Get("date"); customDate != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", customDate, wib); err == nil {
			targetTime = parsed
		}
	}

	payload := ComputePayload(qrType, targetTime)
	rawBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error":"failed to marshal json"}`, http.StatusInternalServerError)
		return
	}

	qrBytes, err := qrcode.Encode(string(rawBytes), qrcode.Medium, 256)
	if err != nil {
		http.Error(w, `{"error":"failed to generate qr"}`, http.StatusInternalServerError)
		return
	}

	format := strings.ToLower(q.Get("format"))
	if format == "" {
		format = strings.ToLower(q.Get("output"))
	}

	switch format {
	case "binary", "image", "png", "raw":
		w.Header().Set("Content-Type", "image/png")
		w.Write(qrBytes)
		return
	case "base64", "b64", "text":
		b64 := base64.StdEncoding.EncodeToString(qrBytes)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(b64))
		return
	default:
		b64 := base64.StdEncoding.EncodeToString(qrBytes)
		resp := QRResponse{
			Payload:   payload,
			RawJSON:   string(rawBytes),
			QRBase64:  b64,
			QRDataURI: "data:image/png;base64," + b64,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
