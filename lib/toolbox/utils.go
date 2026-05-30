package toolbox

import "strconv"

// ToPointer returns a pointer to the given value.
func ToPointer[T any](value T) *T {
	return &value
}

// CheckErr panics if the given error is not nil.
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ParseID convierte un string ID a uint
func ParseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
