package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Ana uygulamayı cmd/server/main.go'dan çalıştır
	fmt.Println("UniNotes uygulaması başlatılıyor...")
	fmt.Println("Ana uygulama cmd/server/main.go'dan çalıştırılıyor...")

	// Çalışma dizinini al
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Çalışma dizini alınamadı: %v", err)
	}

	// cmd/server/main.go'yu çalıştır
	cmd := exec.Command("go", "run", filepath.Join(wd, "cmd", "server", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Uygulama çalıştırılamadı: %v", err)
	}
}
