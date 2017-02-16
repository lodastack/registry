package common

import (
	"testing"
)

func TesRemoveIfContain(t *testing.T) {
	if _, ok := RemoveIfContain([]string{"", "b", ""}, "a"); !ok {
		t.Fatal("case 1 fail not match with expect")
	}
	if _, ok := RemoveIfContain([]string{}, ""); ok {
		t.Fatal("case 2 success not match with expect")
	}
	if _, ok := RemoveIfContain([]string{"a", "b", "c"}, "c"); !ok {
		t.Fatal("case 3 fail not match with expect")
	}
	test := []string{"a", "a", "b"}
	if _, ok := RemoveIfContain(test, "c"); ok {
		t.Fatal("case 4 success not match with expect")
	}

	if _, ok := RemoveIfContain(test, "a"); ok && len(test) != 2 {
		t.Fatal("case 5 not match with expect", ok, len(test))
	}
}
