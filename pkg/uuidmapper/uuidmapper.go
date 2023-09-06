package uuidmapper

import (
	"hash/fnv"
	"math"

	"github.com/gocql/gocql" // Import the Cassandra driver
)

func MapTimeUUIDToInt(uuid gocql.UUID) int {
	uuidStr := uuid.String()
	hash := someHashFunction(uuidStr)
	return hash%(math.MaxInt) + 1 // Map the hash value to the desired integer range
}

func someHashFunction(input string) int {
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return int(hash.Sum32())
}
