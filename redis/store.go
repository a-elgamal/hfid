// Package redis provides a redis implementation for GeneratorStore
package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gitlab.com/alielgamal/hfid"
	"strconv"
)

const prefixKey = "p"
const encodingKey = "e"
const minLengthKey = "m"
const lengthKey = "l"

// GeneratorStore A Struct that wraps a Redis UniversalClient and implements the GeneratorStore interface provided by
// HFID. This implementation utilizes a Hash stored with the generator's key and a HyperLogLog stored with the
// generator's name-hll
type GeneratorStore struct {
	redis.UniversalClient
}

func hllKey(gName string) string {
	return gName + "-hll"
}

// InsertOrGet Implemented by HMGet command then followed by either HMSet or PFCount command.
func (gs GeneratorStore) InsertOrGet(ctx context.Context, g hfid.Generator) (hfid.Generator, int64, error) {
	getCmd := gs.HMGet(ctx, g.Name, prefixKey, encodingKey, minLengthKey, lengthKey)
	if getCmd.Err() != nil {
		return g, 0, getCmd.Err()
	}

	if getCmd.Val()[0] == nil {
		// Insert the Generator
		err := gs.Upsert(ctx, g)
		return g, 0, err
	}
	g.Prefix = getCmd.Val()[0].(string)
	g.Encoding = hfid.Encoding(getCmd.Val()[1].(string))
	ml, err := strconv.ParseInt(getCmd.Val()[2].(string), 10, 8)
	if err != nil || ml < 0 {
		return g, 0, fmt.Errorf("invalid MinLength value '%s' stored for Generator name '%s'", getCmd.Val()[2].(string), g.Name)
	}
	g.MinLength = uint8(ml)
	l, err := strconv.ParseInt(getCmd.Val()[3].(string), 10, 8)
	if err != nil || l < 0 {
		return g, 0, fmt.Errorf("invalid Length value '%s' stored for Generator name '%s'", getCmd.Val()[3].(string), g.Name)
	}
	g.Length = uint8(l)

	countCmd := gs.PFCount(ctx, hllKey(g.Name))
	if countCmd.Err() != nil {
		return g, 0, countCmd.Err()
	}
	return g, countCmd.Val(), nil
}

// Upsert Implemented using HMSet command
func (gs GeneratorStore) Upsert(ctx context.Context, g hfid.Generator) error {
	setCmd := gs.HMSet(ctx, g.Name, prefixKey, g.Prefix, encodingKey, string(g.Encoding), minLengthKey, g.MinLength, lengthKey, g.Length)
	return setCmd.Err()
}

// Add Implemented using PFAdd command
func (gs GeneratorStore) Add(ctx context.Context, hfid int64, gName string) (bool, error) {
	addCmd := gs.PFAdd(ctx, hllKey(gName), hfid)
	return addCmd.Val() == 1, addCmd.Err()
}
