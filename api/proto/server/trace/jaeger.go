package trace

import (
	"therealbroker/internal/exporter"
)

func JaegerRegister() error {
	return exporter.Register()
}
