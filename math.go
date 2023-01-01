package hfid

import "fmt"

// Pow calculates a base int to the power of the exponent. Returns an error if the operation will result an overflow
func Pow(base int, exp int) (int64, error) {
	result := int64(1)
	for i := 0; i < exp; i++ {
		newResult := result * int64(base)
		if base != 0 && newResult/int64(base) != result {
			return 0, fmt.Errorf("%d ^ %d results an overflow", base, exp)
		}
		result = newResult
	}
	return result, nil
}
