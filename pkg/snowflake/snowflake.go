package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

// GenerateSnowflake generates a snowflake ID for a message based on the Twitter snowflake algorithm.
func generateSnowflake() int {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return int(node.Generate().Int64())
}
