package trace

import (
	"log"
	"therealbroker/internal/exporter"
)

func JaegerRegister() error {
	err := exporter.Register()
	if err != nil {
		log.Fatalf("Jaeger failed: %v", err)
		return err
	}
	return nil
}
