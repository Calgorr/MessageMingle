package uuidmapper

import (
	"hash/fnv"
	"math"

	"github.com/gocql/gocql" // Import the Cassandra driver
)

func MapTimeUUIDToInt(uuid gocql.UUID) uint {
	uuidStr := uuid.String()
	hash := uint(someHashFunction(uuidStr))
	return hash%(math.MaxUint) + 1 // Map the hash value to the desired integer range
}

func someHashFunction(input string) uint {
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return uint(hash.Sum32())
}
