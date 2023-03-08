// Package hfid contain the core logic for generating HFIDs. To generate HFIDs, you need to call HFID function with a
// Generator struct and a GeneratorStore implementation of your choosing.
package hfid

import (
	"context"
	"math/rand"
	"time"
)

var defaultRand = *rand.New(rand.NewSource(time.Now().UnixNano()))

// HFID generates a new HFID. If you would like to have deterministic way of generating HFIDs, pass a Rand object,
// otherwise a non-deterministic Rand object will be used. It is recommended to wrap calls to this function with a
// circuit breaker that falls back to a normal UUID when open.
func HFID(ctx context.Context, g Generator, s GeneratorStore, dr ...rand.Rand) (string, error) {
	// Fetch or create the generator
	g, c, err := s.InsertOrGet(ctx, g)
	if err != nil {
		return "", err
	}

	// Checking if we need to increase the length of the generator
	maxC, err := g.countHFIDs()
	if err != nil {
		return "", err
	}
	if c+1 > int64(float64(maxC)*0.5) {
		g.Length++
		err = s.Upsert(ctx, g)
		if err != nil {
			return "", err
		}
	}

	// Prepare a random source
	var r rand.Rand
	if len(dr) == 0 {
		r = defaultRand
	} else {
		r = dr[0]
	}

	// Generate valid HFID
	max, err := g.maxHFID()
	if err != nil {
		return "", err
	}
	var hfid int64
	for {
		hfid = r.Int63n(max + 1)
		isNew, err := s.Add(ctx, hfid, g.Name)
		if err != nil {
			return "", err
		}
		if isNew {
			return g.encodeHFID(hfid)
		}
	}
}
