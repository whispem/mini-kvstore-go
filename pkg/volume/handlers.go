package volume

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/whispem/mini-kvstore-go/pkg/store"
)

var startTime = time.Now()

// AppState holds shared application state
type AppState struct {
	storage *BlobStorage
	mu      sync.RWMutex
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status     string  `json:"status"`
	VolumeID   string  `json:"volume_id"`
	Keys       int     `json:"keys"`
	Segments   int     `json:"segments"`
	TotalMB    float64 `json:"total_mb"`
	UptimeSecs int64   `json:"uptime_secs"`
}

// MetricsResponse represents metrics response
type MetricsResponse struct {
	TotalKeys          int     `json:"total_keys"`
	TotalSegments      int     `json:"total_segments"`
	TotalBytes         uint64  `json:"total_bytes"`
	TotalMB            float64 `json:"total_mb"`
	ActiveSegmentID    int     `json:"active_segment_id"`
	OldestSegmentID    int     `json:"oldest_segment_id"`
	VolumeID           string  `json:"volume_id"`
	UptimeSecs         int64   `json:"uptime_secs"`
	AvgValueSizeBytes  float64 `json:"avg_value_size_bytes"`
}

// CreateRouter creates the HTTP router
func CreateRouter(storage *BlobStorage) *mux.Router {
	state := &AppState{storage: storage}

	r := mux.NewRouter()
	r.HandleFunc("/", state.healthCheck).Methods("GET")
	r.HandleFunc("/health", state.healthCheck).Methods("GET")
	r.HandleFunc("/metrics", state.metrics).Methods("GET")
	r.HandleFunc("/blobs", state.listBlobs).Methods("GET")
	r.HandleFunc("/blobs/{key}", state.putBlob).Methods("POST")
	r.HandleFunc("/blobs/{key}", state.getBlob).Methods("GET")
	r.HandleFunc("/blobs/{key}", state.deleteBlob).Methods("DELETE")

	return r
}

func (s *AppState) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.storage.Stats()
	volumeID := s.storage.VolumeID()
	s.mu.RUnlock()

	response := HealthResponse{
		Status:     "healthy",
		VolumeID:   volumeID,
		Keys:       stats.NumKeys,
		Segments:   stats.NumSegments,
		TotalMB:    stats.TotalMB(),
		UptimeSecs: int64(time.Since(startTime).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *AppState) metrics(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.storage.Stats()
	volumeID := s.storage.VolumeID()
	s.mu.RUnlock()

	avgValueSize := 0.0
	if stats.NumKeys > 0 {
		avgValueSize = float64(stats.TotalBytes) / float64(stats.NumKeys)
	}

	response := MetricsResponse{
		TotalKeys:         stats.NumKeys,
		TotalSegments:     stats.NumSegments,
		TotalBytes:        stats.TotalBytes,
		TotalMB:           stats.TotalMB(),
		ActiveSegmentID:   stats.ActiveSegmentID,
		OldestSegmentID:   stats.OldestSegmentID,
		VolumeID:          volumeID,
		UptimeSecs:        int64(time.Since(startTime).Seconds()),
		AvgValueSizeBytes: avgValueSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *AppState) putBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body: "+err.Error())
		return
	}

	s.mu.Lock()
	meta, err := s.storage.Put(key, data)
	s.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meta)
}

func (s *AppState) getBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.mu.RLock()
	data, err := s.storage.Get(key)
	s.mu.RUnlock()

	if err == store.ErrNotFound {
		writeError(w, http.StatusNotFound, "Blob not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *AppState) deleteBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.mu.Lock()
	err := s.storage.Delete(key)
	s.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AppState) listBlobs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	keys := s.storage.ListKeys()
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
