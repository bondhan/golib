package util

import (
	"testing"
)

func TestNewRandomString_Duplication(t *testing.T) {

	t.Run("length check", func(t *testing.T) {
		num := 1000
		for i := 0; i < num; i++ {
			rdm := NewRandomString(i)

			if len(rdm) != i {
				t.Errorf("string length incorrect, expected: %d got: %d", i, len(rdm))
				return
			}
		}
	})
}
