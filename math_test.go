package hfid

import "testing"

func TestPow(t *testing.T) {
	type args struct {
		base int
		exp  int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"0 ^ 0", args{0, 0}, 1, false},
		{"0 ^ 2", args{0, 2}, 0, false},
		{"1 ^ 0", args{1, 0}, 1, false},
		{"1 ^ 3", args{1, 3}, 1, false},
		{"2 ^ 0", args{2, 0}, 1, false},
		{"3 ^ 1", args{3, 1}, 3, false},
		{"4 ^ 2", args{4, 2}, 16, false},
		{"10 ^ 19", args{10, 19}, 0, true},
		{"-1 ^ 0", args{-1, 0}, 1, false},
		{"-1 ^ 1", args{-1, 1}, -1, false},
		{"-1 ^ 2", args{-1, 2}, 1, false},
		{"-2 ^ 4", args{-2, 4}, 16, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Pow(tt.args.base, tt.args.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pow() got = %v, want %v", got, tt.want)
			}
		})
	}
}
