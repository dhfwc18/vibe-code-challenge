package save

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// MasterSeed is the per-playthrough random seed stored in the save file.
// All sub-seeds are derived deterministically from MasterSeed so that a given
// save always produces the same random outcomes.
type MasterSeed uint64

// NewMasterSeed generates a cryptographically random MasterSeed using the OS
// entropy source. It must be called once per new game, never on load.
func NewMasterSeed() (MasterSeed, error) {
	n, err := rand.Int(rand.Reader, new(big.Int).SetUint64(^uint64(0)))
	if err != nil {
		return 0, fmt.Errorf("seed generation failed: %w", err)
	}
	return MasterSeed(n.Uint64()), nil
}

// DeriveSubSeed returns a deterministic sub-seed for a named subsystem.
// It mixes MasterSeed with the FNV-1a hash of the subsystem name so that each
// subsystem draws from an independent pseudo-random stream.
func (m MasterSeed) DeriveSubSeed(subsystem string) uint64 {
	// FNV-1a 64-bit
	const (
		fnvOffsetBasis uint64 = 14695981039346656037
		fnvPrime       uint64 = 1099511628211
	)
	h := fnvOffsetBasis
	for _, b := range []byte(subsystem) {
		h ^= uint64(b)
		h *= fnvPrime
	}
	// mix with master seed using a simple xorshift step
	x := uint64(m) ^ h
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	x *= 0x2545F4914F6CDD1D
	return x
}
