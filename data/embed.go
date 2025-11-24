package data

import _ "embed"

//go:embed seed.json
var SeedJSON []byte

// GetSeedJSON returns the embedded seed.json content
func GetSeedJSON() []byte {
	return SeedJSON
}
