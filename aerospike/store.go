// Package aerospike provides an aerospike implementation for GeneratorStore
package aerospike

import (
	"context"
	"fmt"
	aero "github.com/aerospike/aerospike-client-go/v6"
	"github.com/hashicorp/go-multierror"
	"gitlab.com/alielgamal/hfid"
	"reflect"
)

const gBin = "g"
const prefixKey = "p"
const encodingKey = "e"
const minLengthKey = "m"
const lengthKey = "l"
const hllBin = "h"

// GeneratorStore A Struct that wraps an Aerospike Client and implements the GeneratorStore interface provided by
// HFID. This implementation utilizes a single set with a bin for the generator's properties and another bin for the
// generator's associated HLL
type GeneratorStore struct {
	client    *aero.Client
	namespace string
	set       string
}

// InsertOrGet Implemented using a single Operate command that creates the generator if it doesn't exist and reads the
// generator properties and the HyperLogLog count estimate.
func (gs GeneratorStore) InsertOrGet(_ context.Context, g hfid.Generator) (hfid.Generator, int64, error) {
	key, aeroErr := aero.NewKey(gs.namespace, gs.set, g.Name)
	merr := multierror.Append(aeroErr)

	// Write the record to Aerospike spike ONLY if it doesn't exist.
	r, aeroErr := gs.client.Operate(nil, key,
		aero.MapPutItemsOp(aero.NewMapPolicyWithFlags(aero.MapOrder.UNORDERED, aero.MapWriteFlagsCreateOnly|aero.MapWriteFlagsNoFail),
			gBin, map[interface{}]interface{}{
				prefixKey:    g.Prefix,
				encodingKey:  g.Encoding,
				minLengthKey: g.MinLength,
				lengthKey:    g.Length,
			}),
		// Read the generator details
		aero.GetBinOp(gBin),
		// Get the HLL Count
		aero.HLLGetCountOp(hllBin),
	)
	merr = multierror.Append(aeroErr, aeroErr)
	if merr.ErrorOrNil() != nil {
		return hfid.Generator{}, 0, merr.ErrorOrNil()
	}

	storedG := r.Bins[gBin].([]interface{})[1].(map[interface{}]interface{})

	p, err := toString(storedG[prefixKey])
	merr = multierror.Append(err)
	g.Prefix = p

	e, err := toString(storedG[encodingKey])
	merr = multierror.Append(merr, err)
	g.Encoding = hfid.Encoding(e)

	ml, err := toInt(storedG[minLengthKey])
	merr = multierror.Append(merr, err)
	g.MinLength = uint8(ml)

	l, err := toInt(storedG[lengthKey])
	merr = multierror.Append(merr, err)
	g.Length = uint8(l)

	switch r.Bins[hllBin].(type) {
	case nil:
		return g, 0, merr.ErrorOrNil()
	case int:
		return g, int64(r.Bins[hllBin].(int)), merr.ErrorOrNil()
	case int64:
		return g, r.Bins[hllBin].(int64), merr.ErrorOrNil()
	default:
		return hfid.Generator{}, 0, multierror.Append(merr, fmt.Errorf("unexpected hll count result %v of type %v", r.Bins[hllBin], reflect.TypeOf(r.Bins[hllBin]))).ErrorOrNil()
	}
}

// Upsert Implemented using a single MapPutItemsOp command
func (gs GeneratorStore) Upsert(_ context.Context, g hfid.Generator) error {
	key, err := aero.NewKey(gs.namespace, gs.set, g.Name)
	merr := multierror.Append(err)

	_, err = gs.client.Operate(nil, key, aero.MapPutItemsOp(aero.DefaultMapPolicy(), gBin,
		map[interface{}]interface{}{
			prefixKey:    g.Prefix,
			encodingKey:  g.Encoding,
			minLengthKey: g.MinLength,
			lengthKey:    g.Length,
		}))
	merr = multierror.Append(err, err)

	return merr.ErrorOrNil()
}

func toString(any interface{}) (string, error) {
	switch any.(type) {
	case string:
		return any.(string), nil
	default:
		return "", fmt.Errorf("expected string type, but found %s", reflect.TypeOf(any))
	}
}

func toInt(any interface{}) (int, error) {
	switch any.(type) {
	case int:
		return any.(int), nil
	default:
		return 0, fmt.Errorf("expected int type, but found %s", reflect.TypeOf(any))
	}
}

// Add Implemented using HLLAddOp command
func (gs GeneratorStore) Add(_ context.Context, hfid int64, gName string) (bool, error) {
	key, err := aero.NewKey(gs.namespace, gs.set, gName)
	merr := multierror.Append(err)

	r, err := gs.client.Operate(nil, key,
		aero.HLLAddOp(aero.DefaultHLLPolicy(), hllBin,
			[]aero.Value{aero.NewLongValue(hfid)}, 16, 4))
	merr = multierror.Append(merr, err)

	if merr.ErrorOrNil() != nil {
		return false, merr.ErrorOrNil()
	}

	switch r.Bins[hllBin].(type) {
	case int:
		return r.Bins[hllBin].(int) > 0, merr.ErrorOrNil()
	default:
		return false, multierror.Append(merr, fmt.Errorf("hll Add Operation didn't return an int:  %s", r.Bins[hllBin])).ErrorOrNil()
	}
}
