package hfid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockGeneratorStore struct {
	mock.Mock
}

func (mgs *MockGeneratorStore) InsertOrGet(ctx context.Context, g Generator) (Generator, int64, error) {
	args := mgs.MethodCalled("InsertOrGet", ctx, g)
	return args.Get(0).(Generator), args.Get(1).(int64), args.Error(2)
}

func (mgs *MockGeneratorStore) Upsert(ctx context.Context, g Generator) error {
	args := mgs.MethodCalled("Upsert", ctx, g)
	return args.Error(0)
}

func (mgs *MockGeneratorStore) Add(ctx context.Context, hfid int64, gName string) (bool, error) {
	args := mgs.MethodCalled("Add", ctx, hfid, gName)
	return args.Bool(0), args.Error(1)
}

func NewMockGeneratorStore(t *testing.T) *MockGeneratorStore {
	result := MockGeneratorStore{}
	result.Test(t)
	return &result
}
