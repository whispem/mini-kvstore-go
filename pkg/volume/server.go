package volume

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type VolumeServer struct {
	dir string
}

func NewVolumeServer(dir string) *VolumeServer {
	return &VolumeServer{dir: dir}
}

func (vs *VolumeServer) Start(addr string) error {
	fmt.Println("Starting volume server")
	http.HandleFunc("/upload", vs.uploadHandler)
	http.HandleFunc("/files/", vs.fileHandler)
	return http.ListenAndServe(addr, nil)
}

func (vs *VolumeServer) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := os.Create(filepath.Join(vs.dir, header.Filename))
	if err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = f.ReadFrom(file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File %s uploaded successfully!", header.Filename)
}

func (vs *VolumeServer) fileHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	http.ServeFile(w, r, filepath.Join(vs.dir, filename))
}

func (vs *VolumeServer) Stop() {
	log.Println("Stopping volume server")
}
