package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func computePayload(qrType string, refTime time.Time) QRPayload {
	// ponytail: offset hari relatif 2026-10-14 WIB, upgrade DB sequence jika butuh skip hari libur
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

func handleQR(w http.ResponseWriter, r *http.Request) {
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

	payload := computePayload(qrType, targetTime)
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

	if q.Get("format") == "image" {
		w.Header().Set("Content-Type", "image/png")
		w.Write(qrBytes)
		return
	}

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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /qr", handleQR)
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /", handleQR) // default fallback

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
