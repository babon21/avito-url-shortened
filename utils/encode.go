package utils

const (
	base         uint64 = 62
	characterSet        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func ToBase62(num uint64) string {
	encoded := ""

	for num > 0 {
		r := num % base
		num /= base
		encoded = string(characterSet[r]) + encoded
	}
	return encoded
}
