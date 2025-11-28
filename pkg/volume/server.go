package volume

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	Addr                string
	VolumeID            string
	DataDir             string
	CompactionThreshold int
	CompactionInterval  time.Duration
	httpServer          *http.Server
}

func NewServer(addr, volumeID, dataDir string, compactionThreshold int, compactionIntervalSecs int) *Server {
	return &Server{
		Addr:                addr,
		VolumeID:            volumeID,
		DataDir:             dataDir,
		CompactionThreshold: compactionThreshold,
		CompactionInterval:  time.Duration(compactionIntervalSecs) * time.Second,
	}
}

func (s *Server) Start() error {
	
	if err := os.MkdirAll(s.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}


	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.httpServer = &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	
	go func() {
		ticker := time.NewTicker(s.CompactionInterval)
		defer ticker.Stop()
		for range ticker.C {
			log.Printf("[volume %s] Running compaction (threshold=%d)...", s.VolumeID, s.CompactionThreshold)
			
		}
	}()

	log.Printf("Server listening on %s\n", s.Addr)
	return s.httpServer.ListenAndServe()
}

func StartVolumeServer(addr, volumeID, dataDir string, compactionThreshold, compactionIntervalSecs int) error {
	s := NewServer(addr, volumeID, dataDir, compactionThreshold, compactionIntervalSecs)
	return s.Start()
}
