package hfid

import (
	"context"
	"fmt"
	"strings"
)

// Generator Specified how HFIDs should be generated. Please use NewGenerator to create instances of this struct to
// ensure correctness of the values passed.
type Generator struct {
	Name      string
	Prefix    string
	Encoding  Encoding
	MinLength uint8
	Length    uint8
}

// GeneratorStore interface to store and update Generator Instances
type GeneratorStore interface {
	// InsertOrGet a Generator. The function returns the Generator that has been created or found in the store and an
	// estimate number of HFIDs that have been generated using this generator using the associated hyperloglog (0 if the
	// generator was created)
	InsertOrGet(ctx context.Context, g Generator) (Generator, int64, error)

	// Upsert a Generator. If an existing Generator with the same name was found, update it without changing the
	// hyperloglog attached to it. Otherwise, Insert a new Generator with a new hyperloglog.
	Upsert(ctx context.Context, g Generator) (Generator, error)

	// Add hfid to the hyperloglog associated with the generator named gName. Return true if the hyperloglog was changed
	// , false otherwise.
	Add(ctx context.Context, hfid int64, gName string) (bool, error)
}

// NewGenerator creates a new Generator after validating the arguments
func NewGenerator(name string, prefix string, e Encoding, minLength uint8, length uint8) (*Generator, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	if err := e.Valid(); err != nil {
		return nil, fmt.Errorf("invalid Encoding ('%s'): %s", e, err)
	}

	if length == 0 {
		return nil, fmt.Errorf("length cannot be zero")
	}
	if length < minLength {
		return nil, fmt.Errorf("length '%d' cannot be less than MinLength '%d'", length, minLength)
	}

	result := Generator{name, prefix, e, minLength, length}

	if _, err := result.maxHFID(); err != nil {
		return nil, fmt.Errorf("encoding '%s' with Length %d would result overflow. This Generator cannot be used any more", e, length)
	}

	return &result, nil
}

func (it Generator) maxHFID() (int64, error) {
	maxPlus1, err := Pow(len(it.Encoding), int(it.Length))
	if err != nil {
		return 0, err
	}
	return maxPlus1 - 1, err
}

func (it Generator) countHFIDs() (int64, error) {
	max, err := it.maxHFID()

	if err != nil {
		return 0, err
	}
	return max + 1, nil
}

func (it Generator) encodeHFID(n int64) (string, error) {
	maxN, err := it.maxHFID()
	if n > maxN {
		return "", fmt.Errorf("%d is bigger than %d which is the maximum number that can be encoded with encoding '%s' with %d characters", n, maxN, it.Encoding, it.Length)
	}

	result, err := it.Encoding.Encode(n)
	if err != nil {
		return "", err
	}

	if len(result) < int(it.Length) {
		result = strings.Repeat(string(it.Encoding[0]), int(it.Length)-len(result)) + result
	}

	return it.Prefix + result, nil
}

func (it Generator) decodeHFID(hfid string) (int64, error) {
	if !strings.HasPrefix(hfid, it.Prefix) {
		return 0, fmt.Errorf("cannot decode HFID '%s' of type '%s' must have the prefix '%s'", hfid, it.Name, it.Prefix)
	}

	// Remove the prefix
	hfid = hfid[len(it.Prefix):]

	// Validate the length
	if len(hfid) > int(it.Length) || len(hfid) < int(it.MinLength) {
		return 0, fmt.Errorf("cannot decode HFID '%s' of type '%s' its length is not between [%d, %d]", hfid, it.Name, it.MinLength, it.Length)
	}

	return it.Encoding.Decode(hfid)
}
