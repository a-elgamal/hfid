package redis

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"gitlab.com/alielgamal/hfid"
	"strconv"
	"testing"
)

func TestClient_InsertOrGet(t *testing.T) {
	insertOrGetWithFixtures := func(g hfid.Generator, fixturesF func(*miniredis.Miniredis)) (hfid.Generator, int64, error) {
		mr := miniredis.RunT(t)
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
		gs := GeneratorStore{uc}

		fixturesF(mr)

		return gs.InsertOrGet(context.Background(), g)
	}

	t.Run("Inserts a new generator when it doesn't exist", func(t *testing.T) {
		g := hfid.Generator{
			Name:      t.Name(),
			Prefix:    t.Name()[0:2],
			Encoding:  hfid.NumericEncoding,
			MinLength: 1,
			Length:    2,
		}
		g2, c, err := insertOrGetWithFixtures(g, func(_ *miniredis.Miniredis) {})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), c)
		assert.Equal(t, g, g2)
	})

	t.Run("Returns existing generator", func(t *testing.T) {
		g := hfid.Generator{
			Name:      t.Name(),
			Prefix:    t.Name()[0:2],
			Encoding:  hfid.NumericEncoding,
			MinLength: 1,
			Length:    2,
		}

		newG := hfid.Generator{Name: t.Name()}

		g2, c, err := insertOrGetWithFixtures(newG, func(mr *miniredis.Miniredis) {
			mr.HSet(g.Name,
				prefixKey, t.Name()[0:2],
				encodingKey, hfid.NumericEncoding,
				minLengthKey, strconv.FormatUint(uint64(g.MinLength), 10),
				lengthKey, strconv.FormatUint(uint64(g.Length), 10),
			)
			_, err := mr.PfAdd(hllKey(g.Name), "1", "2", "3")
			assert.NoError(t, err)
		})

		assert.NoError(t, err)
		assert.Equal(t, int64(3), c)
		assert.Equal(t, g, g2)
	})

	t.Run("Fails if the existing generator data is corrupt", func(t *testing.T) {
		newG := hfid.Generator{Name: t.Name()}

		tcs := []struct {
			name      string
			minLength string
			length    string
		}{
			{"minLength negative", "-1", "0"},
			{"minLength overflow", "1000", "0"},
			{"minLength invalid", "a", "0"},
			{"length negative", "0", "-1"},
			{"length overflow", "0", "1000"},
			{"length invalid", "0", "a"},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := insertOrGetWithFixtures(newG, func(mr *miniredis.Miniredis) {
					mr.HSet(newG.Name,
						prefixKey, t.Name()[0:2],
						encodingKey, hfid.NumericEncoding,
						minLengthKey, tc.minLength,
						lengthKey, tc.length,
					)
				})
				assert.Error(t, err)
			})
		}
	})

	t.Run("Fails if Redis client fails", func(t *testing.T) {
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{""}})
		gs := GeneratorStore{uc}

		_, _, err := gs.InsertOrGet(context.Background(), hfid.Generator{})
		assert.Error(t, err)
	})
}

func TestGeneratorStore_Upsert(t *testing.T) {
	g := hfid.Generator{
		Name:      t.Name(),
		Prefix:    t.Name()[0:2],
		Encoding:  hfid.NumericEncoding,
		MinLength: 1,
		Length:    2,
	}

	assertUpsertWithFixtures := func(fixturesF func(*miniredis.Miniredis)) {
		mr := miniredis.RunT(t)
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
		gs := GeneratorStore{uc}

		fixturesF(mr)

		g2, err := gs.Upsert(context.Background(), g)
		assert.NoError(t, err)
		assert.Equal(t, g, g2)
		assert.Equal(t, g.Prefix, mr.HGet(g.Name, prefixKey))
		assert.Equal(t, string(g.Encoding), mr.HGet(g.Name, encodingKey))
		assert.Equal(t, strconv.Itoa(int(g.MinLength)), mr.HGet(g.Name, minLengthKey))
		assert.Equal(t, strconv.Itoa(int(g.Length)), mr.HGet(g.Name, lengthKey))
	}

	t.Run("Inserts a new Generator if none existed", func(t *testing.T) {
		assertUpsertWithFixtures(func(_ *miniredis.Miniredis) {})
	})

	t.Run("Updates existing generator", func(t *testing.T) {
		assertUpsertWithFixtures(func(mr *miniredis.Miniredis) {
			mr.HSet(g.Name, prefixKey, "", encodingKey, "", minLengthKey, "0", lengthKey, "0")
		})
	})

	t.Run("Fails if Redis client fails", func(t *testing.T) {
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{""}})
		gs := GeneratorStore{uc}

		_, err := gs.Upsert(context.Background(), hfid.Generator{})
		assert.Error(t, err)
	})
}

func TestGeneratorStore_Add(t *testing.T) {
	gName := "generator"
	id := int64(1)

	assertAddWithFixtures := func(t *testing.T, fixtureF func(*miniredis.Miniredis), expectedReturn bool, expectedCount int) {
		mr := miniredis.RunT(t)
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
		gs := GeneratorStore{uc}

		fixtureF(mr)

		u, err := gs.Add(context.Background(), id, gName)
		assert.NoError(t, err)
		assert.Equal(t, expectedReturn, u)
		c, err := mr.PfCount(hllKey(gName))
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, c)
	}

	t.Run("Creates HLL When it doesn't exist", func(t *testing.T) {
		assertAddWithFixtures(t, func(_ *miniredis.Miniredis) {}, true, 1)
	})

	t.Run("Updates existing HLL", func(t *testing.T) {
		assertAddWithFixtures(t, func(mr *miniredis.Miniredis) {
			_, err := mr.PfAdd(hllKey(gName), "ali")
			assert.NoError(t, err)
		}, true, 2)
	})

	t.Run("Returns false when the element was previously added to HLL", func(t *testing.T) {
		assertAddWithFixtures(t, func(mr *miniredis.Miniredis) {
			_, err := mr.PfAdd(hllKey(gName), strconv.FormatInt(id, 10))
			assert.NoError(t, err)
		}, false, 1)
	})

	t.Run("Fails if Redis client fails", func(t *testing.T) {
		uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{""}})
		gs := GeneratorStore{uc}

		_, err := gs.Add(context.Background(), id, gName)
		assert.Error(t, err)
	})
}
