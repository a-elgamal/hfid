package aerospike

import (
	"context"
	aero "github.com/aerospike/aerospike-client-go/v6"
	"github.com/stretchr/testify/assert"
	"gitlab.com/alielgamal/hfid"
	"log"
	"os"
	"strings"
	"testing"
)

var aerospikeClient *aero.Client

func TestMain(m *testing.M) {
	host, set := os.LookupEnv("AEROSPIKE_HOST")
	if !set {
		host = "localhost"
	}
	c, err := aero.NewClient(host, 3000)
	if err != nil {
		log.Fatal(err)
	}

	result := func() int {
		aerospikeClient = c
		defer c.Close()
		return m.Run()
	}()

	os.Exit(result)
}

func prepareStore(t *testing.T) GeneratorStore {
	namespace := "test"
	setName := strings.ReplaceAll(t.Name(), "/", "_")
	const setLength = 60
	startIndex := 0
	if len(setName) > setLength {
		startIndex = len(setName) - setLength
	}
	setName = setName[startIndex:]
	_ = aerospikeClient.Truncate(nil, namespace, setName, nil)

	return GeneratorStore{
		Client:    aerospikeClient,
		Namespace: namespace,
		Set:       setName,
	}
}

func TestGeneratorStore_InsertOrGet(t *testing.T) {
	name := "test"
	prefix := "t-"
	encoding := hfid.Encoding(hfid.DefaultEncoding)
	minLength := uint8(1)
	length := uint8(2)

	t.Run("new generator is added zero count", func(t *testing.T) {
		gs := prepareStore(t)
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)
		foundG, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.NoError(t, err)
		assert.Equal(t, *g, foundG)
		assert.Equal(t, int64(0), c)
	})

	t.Run("existing generator is returned", func(t *testing.T) {
		existingPrefix := "t2-"
		existingEncoding := hfid.Encoding(hfid.NumericEncoding)
		existingMinLength := uint8(2)
		existingLength := uint8(3)
		existingG, err := hfid.NewGenerator(name, existingPrefix, existingEncoding, existingMinLength, existingLength)
		assert.NoError(t, err)

		gs := prepareStore(t)
		_, c, err := gs.InsertOrGet(context.Background(), *existingG)
		assert.NoError(t, err)
		_, err = gs.Add(context.Background(), 12345, name)
		assert.NoError(t, err)

		// Attempt to add another generator with the same name
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)

		// The returned generator should be the existing one not the new one
		foundG, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.NoError(t, err)
		assert.Equal(t, foundG, *existingG)
		assert.Equal(t, int64(1), c)
	})

	t.Run("returns an error when aerospike fails", func(t *testing.T) {
		gs := prepareStore(t)
		gs.Namespace = "invalid_namespace"
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)
		_, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.Error(t, err)
		assert.Equal(t, int64(0), c)
	})
}

func TestGeneratorStore_Upsert(t *testing.T) {
	name := "test"
	prefix := "t-"
	encoding := hfid.Encoding(hfid.DefaultEncoding)
	minLength := uint8(1)
	length := uint8(2)

	t.Run("creates a new generator if it doesn't exist", func(t *testing.T) {
		gs := prepareStore(t)
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)
		err = gs.Upsert(context.Background(), *g)
		assert.NoError(t, err)

		newPrefix := "t2-"
		newEncoding := hfid.Encoding(hfid.NumericEncoding)
		newMinLength := uint8(2)
		newLength := uint8(3)
		newG, err := hfid.NewGenerator(name, newPrefix, newEncoding, newMinLength, newLength)
		assert.NoError(t, err)

		foundG, _, err := gs.InsertOrGet(context.Background(), *newG)
		assert.NoError(t, err)
		assert.Equal(t, *g, foundG)
	})

	t.Run("updates existing generator", func(t *testing.T) {
		existingPrefix := "t2-"
		existingEncoding := hfid.Encoding(hfid.NumericEncoding)
		existingMinLength := uint8(2)
		existingLength := uint8(3)
		existingG, err := hfid.NewGenerator(name, existingPrefix, existingEncoding, existingMinLength, existingLength)
		assert.NoError(t, err)

		gs := prepareStore(t)
		_, c, err := gs.InsertOrGet(context.Background(), *existingG)
		assert.NoError(t, err)
		_, err = gs.Add(context.Background(), 12345, name)
		assert.NoError(t, err)

		// Upsert
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)
		err = gs.Upsert(context.Background(), *g)
		assert.NoError(t, err)

		// The returned generator should be the updated one with the old hll count
		foundG, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.NoError(t, err)
		assert.Equal(t, foundG, *g)
		assert.Equal(t, int64(1), c)
	})

	t.Run("returns an error when aerospike fails", func(t *testing.T) {
		gs := prepareStore(t)
		gs.Namespace = "invalid_namespace"
		g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
		assert.NoError(t, err)
		err = gs.Upsert(context.Background(), *g)
		assert.Error(t, err)
	})
}

func TestGeneratorStore_Add(t *testing.T) {
	name := "test"
	prefix := "t-"
	encoding := hfid.Encoding(hfid.DefaultEncoding)
	minLength := uint8(1)
	length := uint8(2)
	g, err := hfid.NewGenerator(name, prefix, encoding, minLength, length)
	assert.NoError(t, err)

	addFirstHFID := func(t *testing.T, gs GeneratorStore, id int64) {
		err = gs.Upsert(context.Background(), *g)
		assert.NoError(t, err)

		changed, err := gs.Add(context.Background(), id, name)
		assert.NoError(t, err)
		assert.True(t, changed)
	}

	t.Run("returns true when the element is unique", func(t *testing.T) {
		gs := prepareStore(t)

		addFirstHFID(t, gs, 0)
		changed, err := gs.Add(context.Background(), 1, name)
		assert.NoError(t, err)
		assert.True(t, changed)

		_, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), c)
	})

	t.Run("returns false when the element is duplicated", func(t *testing.T) {
		gs := prepareStore(t)

		addFirstHFID(t, gs, 0)
		changed, err := gs.Add(context.Background(), 0, name)
		assert.NoError(t, err)
		assert.False(t, changed)

		_, c, err := gs.InsertOrGet(context.Background(), *g)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), c)
	})

	t.Run("returns an error when the generator doesn't exist", func(t *testing.T) {
		gs := prepareStore(t)
		gs.Namespace = "invalid_namespace"
		_, err := gs.Add(context.Background(), 0, "non-existent")
		assert.Error(t, err)
	})

	t.Run("returns an error when aerospike fails", func(t *testing.T) {
		gs := prepareStore(t)
		gs.Namespace = "invalid_namespace"
		_, err := gs.Add(context.Background(), 0, name)
		assert.Error(t, err)
	})
}
