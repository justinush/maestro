package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/justinush/maestro/pkg/run"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	rt, err := loadRuntime()
	if err != nil {
		log.Fatal(err)
	}
	if def := rt.Definition(); def != nil {
		fmt.Printf("loaded workflow %q (version %q)\n", def.ID, def.Version)
	}

	svc := NewKYCService(rt, run.NewMemoryStore(), NewApplicantStore())
	srv := &Server{svc: svc}

	fmt.Printf("kyc-backend listening on %s\n", addr)
	fmt.Println("  POST /kyc/start")
	fmt.Println("  GET  /kyc/{runID}")
	fmt.Println("  POST /kyc/{runID}/profile")
	fmt.Println("  POST /kyc/{runID}/document")
	fmt.Println("  POST /kyc/{runID}/review")

	log.Fatal(http.ListenAndServe(addr, srv.routes()))
}

func newID(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:])
}
