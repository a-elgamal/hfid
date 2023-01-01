package hfid

import (
	"testing"
)

func TestEncoding_Valid(t *testing.T) {
	tests := []struct {
		name    string
		e       Encoding
		wantErr bool
	}{
		{"empty Encoding is not valid", "", true},
		{"1 character Encoding is not valid", "", true},
		{"2 characters Encoding is not valid", "", true},
		{"Encoding with duplicate characters is not valid", "", true},
		{"3 unique characters Encoding is valid", "123", false},
		{"default characters Encoding is valid", "0123456789abcdefghijklmnopqrstuvwxyz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_encoding_Encode(t *testing.T) {
	type args struct {
		number int64
	}
	tests := []struct {
		name    string
		e       Encoding
		args    args
		want    string
		wantErr bool
	}{
		{"Fails to encode with invalid Encoding", "", args{int64(0)}, "", true},
		{"Fails to encode negative number", "abc", args{int64(-1)}, "", true},
		{"Correctly encodes 0", "abc", args{int64(0)}, "a", false},
		{"Correctly encodes 1 in a 3-chars Encoding", "abc", args{int64(1)}, "b", false},
		{"Correctly encodes 2 in a 3-chars Encoding", "abc", args{int64(2)}, "c", false},
		{"Correctly encodes 3 in a 3-chars Encoding", "abc", args{int64(3)}, "ba", false},
		{"Correctly encodes 4 in a 3-chars Encoding", "abc", args{int64(4)}, "bb", false},
		{"Correctly encodes 5 in a 3-chars Encoding", "abc", args{int64(5)}, "bc", false},
		{"Correctly encodes 30 in a 4-chars Encoding", "abcd", args{int64(30)}, "bdc", false},
		{"Correctly encodes 63 in a 4-chars Encoding", "abcd", args{int64(63)}, "ddd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Encode(tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encoding_Decode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		e       Encoding
		args    args
		want    int64
		wantErr bool
	}{
		{"Fails when using invalid encoding", "", args{"a"}, int64(0), true},
		{"Fails when overflow occurs", "0123456789", args{"99999999999999999999"}, int64(0), true},
		{"Fails when the string has an invalid character", "abc", args{"d"}, int64(0), true},
		{"Fails when the string has duplicate character", "aac", args{"a"}, int64(0), true},
		{"Correctly decodes 0", "abc", args{"a"}, int64(0), false},
		{"Correctly decodes 1 in a 3-chars Encoding", "abc", args{"b"}, int64(1), false},
		{"Correctly decodes with zeros prefix", "abc", args{"aaaab"}, int64(1), false},
		{"Correctly decodes 2 in a 3-chars Encoding", "abc", args{"c"}, int64(2), false},
		{"Correctly decodes 3 in a 3-chars Encoding", "abc", args{"ba"}, int64(3), false},
		{"Correctly decodes 4 in a 3-chars Encoding", "abc", args{"bb"}, int64(4), false},
		{"Correctly decodes 5 in a 3-chars Encoding", "abc", args{"bc"}, int64(5), false},
		{"Correctly decodes 30 in a 4-chars Encoding", "abcd", args{"bdc"}, int64(30), false},
		{"Correctly decodes 63 in a 4-chars Encoding", "abcd", args{"ddd"}, int64(63), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Decode(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
