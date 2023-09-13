package snowflake

import (
	"context"
	"therealbroker/internal/exporter"

	"github.com/bwmarrin/snowflake"
	"go.opentelemetry.io/otel"
)

// GenerateSnowflake generates a snowflake ID for a message based on the Twitter snowflake algorithm.
func GenerateSnowflake(ctx context.Context) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "GenerateSnowflake method")
	defer globalSpan.End()
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return int(node.Generate().Int64())
}
