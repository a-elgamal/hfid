package hfid

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockGeneratorStore struct {
	mock.Mock
}

func (mgs *MockGeneratorStore) InsertOrGet(g Generator) (Generator, int64, error) {
	args := mgs.MethodCalled("InsertOrGet", g)
	return args.Get(0).(Generator), args.Get(1).(int64), args.Error(2)
}

func (mgs *MockGeneratorStore) Upsert(g Generator) (Generator, error) {
	args := mgs.MethodCalled("Upsert", g)
	return args.Get(0).(Generator), args.Error(1)
}

func (mgs *MockGeneratorStore) Add(hfid int64, gName string) (bool, error) {
	args := mgs.MethodCalled("Add", hfid, gName)
	return args.Bool(0), args.Error(1)
}

func NewMockGeneratorStore(t *testing.T) *MockGeneratorStore {
	result := MockGeneratorStore{}
	result.Test(t)
	return &result
}
