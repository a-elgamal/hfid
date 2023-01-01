package hfid

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHFID(t *testing.T) {
	g, err := NewGenerator("a", "", NumericEncoding, 0, 1)
	assert.NoError(t, err)

	t.Run("fails when store returns an error", func(t *testing.T) {
		t.Run("InsertOrGet", func(t *testing.T) {
			mgs := NewMockGeneratorStore(t)
			mgs.On("InsertOrGet", *g).Return(*g, int64(0), fmt.Errorf("mock error"))
			_, err := HFID(*g, mgs)
			assert.Error(t, err)
			mgs.AssertExpectations(t)
		})

		t.Run("Add", func(t *testing.T) {
			mgs := NewMockGeneratorStore(t)
			mgs.On("InsertOrGet", *g).Return(*g, int64(0), nil)
			mgs.On("Add", mock.Anything, g.Name).Return(false, fmt.Errorf("mock error"))
			_, err := HFID(*g, mgs)
			assert.Error(t, err)
			mgs.AssertExpectations(t)
		})

		t.Run("Upsert", func(t *testing.T) {
			mgs := NewMockGeneratorStore(t)
			mgs.On("InsertOrGet", *g).Return(*g, int64(6), nil)
			mgs.On("Upsert", mock.Anything).Return(*g, fmt.Errorf("mock error"))
			_, err := HFID(*g, mgs)
			assert.Error(t, err)
			mgs.AssertExpectations(t)
		})
	})

	t.Run("fails when generator returns an error", func(t *testing.T) {
		wrongG := Generator{Encoding: NumericEncoding, Length: 20}
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", wrongG).Return(wrongG, int64(0), nil)

		_, err := HFID(wrongG, mgs)
		assert.Error(t, err)
		mgs.AssertExpectations(t)
	})

	t.Run("Generates HFIDs deterministically when a random source is passed", func(t *testing.T) {
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", *g).Return(*g, int64(0), nil)
		mgs.On("Add", int64(0), g.Name).Return(true, nil)

		hfid1, err := HFID(*g, mgs, *rand.New(rand.NewSource(1)))
		assert.NoError(t, err)

		hfid2, err := HFID(*g, mgs, *rand.New(rand.NewSource(1)))
		assert.Equal(t, hfid1, hfid2)
		assert.NoError(t, err)
		mgs.AssertExpectations(t)
	})

	t.Run("Generates HFIDs indeterministically when no random source is passed", func(t *testing.T) {
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", *g).Return(*g, int64(0), nil)
		mgs.On("Add", mock.Anything, g.Name).Return(true, nil)

		hfid, err := HFID(*g, mgs)
		assert.NoError(t, err)
		for i := 0; i < 100; i++ {
			nextHFID, err := HFID(*g, mgs)
			assert.NoError(t, err)
			if hfid != nextHFID {
				mgs.AssertExpectations(t)
				return
			}
		}
		assert.Fail(t, "HFID seems to be generated deterministically when no random source is passed")
	})

	t.Run("Generates HFID using existing generator length", func(t *testing.T) {
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", *g).Return(Generator{"a", "", NumericEncoding, 0, 3}, int64(0), nil)
		mgs.On("Add", mock.Anything, g.Name).Return(true, nil)

		hfid, err := HFID(*g, mgs)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(hfid))
		mgs.AssertExpectations(t)
	})

	t.Run("Retries HFID generation when a duplicate HFID is encountered", func(t *testing.T) {
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", *g).Return(*g, int64(0), nil)
		mgs.On("Add", int64(0), g.Name).Return(false, nil).Once()
		mgs.On("Add", int64(1), g.Name).Return(true, nil).Once()

		hfid, err := HFID(*g, mgs, *rand.New(rand.NewSource(1)))
		assert.NoError(t, err)
		assert.Equal(t, "1", hfid)
		mgs.AssertExpectations(t)
	})

	t.Run("Extends the length of HFID when 50% of HFIDs at the current length have been generated", func(t *testing.T) {
		mgs := NewMockGeneratorStore(t)
		mgs.On("InsertOrGet", *g).Return(*g, int64(6), nil)
		mgs.On("Add", int64(10), g.Name).Return(true, nil).Once()
		newG := *g
		newG.Length++
		mgs.On("Upsert", newG).Return(newG, nil)

		hfid, err := HFID(*g, mgs, *rand.New(rand.NewSource(1)))
		assert.Equal(t, "10", hfid)
		assert.NoError(t, err)
		mgs.AssertExpectations(t)
	})
}
