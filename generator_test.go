package hfid

import (
	"reflect"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	type args struct {
		Name      string
		Prefix    string
		Encoding  Encoding
		MinLength uint8
		Length    uint8
	}
	tests := []struct {
		name    string
		args    args
		want    *Generator
		wantErr bool
	}{
		{"fails with invalid Encoding", args{Name: "a", Encoding: "a", Length: 1}, nil, true},
		{"fails with zero Length", args{Name: "a", Encoding: "abc", Length: 0}, nil, true},
		{"fails if Length is less than MinLength", args{Name: "a", Encoding: "abc", MinLength: 2, Length: 1}, nil, true},
		{"fails if Encoding and Length are too large", args{Name: "a", Encoding: NumericEncoding, MinLength: 20, Length: 20}, nil, true},
		{"fails if Name is empty", args{Name: " ", Encoding: NumericEncoding, MinLength: 2, Length: 2}, nil, true},
		{"creates Generator with passed parameters", args{"a", "a_", "abc", 1, 3}, &Generator{"a", "a_", "abc", 1, 3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGenerator(tt.args.Name, tt.args.Prefix, tt.args.Encoding, tt.args.MinLength, tt.args.Length)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGenerator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_maxHFID(t *testing.T) {
	type fields struct {
		Name      string
		Prefix    string
		Encoding  Encoding
		MinLength uint8
		Length    uint8
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{"Encoding Length - 1 when Length is 1", fields{Encoding: "abc", Length: 1}, 2, false},
		{"Otherwise encoding Length ^ Length - 1", fields{Encoding: "abcd", Length: 2}, 15, false},
		{"Fails when there is an overflow", fields{Encoding: NumericEncoding, Length: 19}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := Generator{
				Name:      tt.fields.Name,
				Prefix:    tt.fields.Prefix,
				Encoding:  tt.fields.Encoding,
				MinLength: tt.fields.MinLength,
				Length:    tt.fields.Length,
			}
			got, err := it.maxHFID()
			if (err != nil) != tt.wantErr {
				t.Errorf("maxHFID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("maxHFID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_countHFIDs(t *testing.T) {
	type fields struct {
		Name      string
		Prefix    string
		Encoding  Encoding
		MinLength uint8
		Length    uint8
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{"Encoding Length when Length is 1", fields{Encoding: "abcd", Length: 1}, 4, false},
		{"Encoding Length when Length is 1", fields{Encoding: "abc", Length: 1}, 3, false},
		{"Fails when there is an overflow", fields{Encoding: NumericEncoding, Length: 19}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := Generator{
				Name:      tt.fields.Name,
				Prefix:    tt.fields.Prefix,
				Encoding:  tt.fields.Encoding,
				MinLength: tt.fields.MinLength,
				Length:    tt.fields.Length,
			}
			got, err := it.countHFIDs()
			if (err != nil) != tt.wantErr {
				t.Errorf("countHFIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("countHFIDs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_encodeHFID(t *testing.T) {
	type fields struct {
		Name      string
		Prefix    string
		Encoding  Encoding
		MinLength uint8
		Length    uint8
	}
	type args struct {
		n int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"Zero", fields{Encoding: NumericEncoding, Prefix: "", Length: 1}, args{0}, "0", false},
		{"Single Character", fields{Encoding: NumericEncoding, Prefix: "a-", Length: 1}, args{4}, "a-4", false},
		{"Single Character but higher length", fields{Encoding: NumericEncoding, Prefix: "", Length: 4}, args{4}, "0004", false},
		{"Multiple Characters", fields{Encoding: NumericEncoding, Prefix: "", Length: 9}, args{123456789}, "123456789", false},
		{"Fails when the number is negative", fields{Encoding: NumericEncoding, Prefix: "", Length: 2}, args{-1}, "", true},
		{"Fails when the number is too large", fields{Encoding: NumericEncoding, Prefix: "", Length: 2}, args{100}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := Generator{
				Name:      tt.fields.Name,
				Prefix:    tt.fields.Prefix,
				Encoding:  tt.fields.Encoding,
				MinLength: tt.fields.MinLength,
				Length:    tt.fields.Length,
			}
			got, err := it.encodeHFID(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeHFID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("encodeHFID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_decodeHFID(t *testing.T) {
	type fields struct {
		Name      string
		Prefix    string
		Encoding  Encoding
		MinLength uint8
		Length    uint8
	}
	type args struct {
		hfid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{"Fails if the HFID is missing the prefix", fields{Encoding: NumericEncoding, Prefix: "a-", Length: 1}, args{"01"}, 0, true},
		{"Fails if a prefixed HFID length is too short", fields{Encoding: NumericEncoding, Prefix: "a-", MinLength: 2}, args{"a-0"}, 0, true},
		{"Fails if a non-prefixed HFID length is too short", fields{Encoding: NumericEncoding, Prefix: "", MinLength: 2}, args{"0"}, 0, true},
		{"Fails if a prefixed HFID length is too long", fields{Encoding: NumericEncoding, Prefix: "a-", Length: 1}, args{"a-01"}, 0, true},
		{"Fails if a non-prefixed HFID length is too long", fields{Encoding: NumericEncoding, Prefix: "", Length: 1}, args{"01"}, 0, true},
		{"Fails if decoding fails", fields{Encoding: NumericEncoding, Prefix: "", Length: 1}, args{"a"}, 0, true},
		{"Otherwise, correctly decodes HFID", fields{Encoding: NumericEncoding, Prefix: "123", Length: 1}, args{"1233"}, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := Generator{
				Name:      tt.fields.Name,
				Prefix:    tt.fields.Prefix,
				Encoding:  tt.fields.Encoding,
				MinLength: tt.fields.MinLength,
				Length:    tt.fields.Length,
			}
			got, err := it.decodeHFID(tt.args.hfid)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeHFID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decodeHFID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
