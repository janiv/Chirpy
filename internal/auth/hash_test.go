package auth

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	cases := []struct {
		key string
	}{
		{
			key: "reallydifficultword",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			res, err := HashPassword(c.key)
			if err != nil {
				t.Errorf("HashPassword returned an error")
				return
			}
			comp_err := bcrypt.CompareHashAndPassword([]byte(res), []byte(c.key))
			if comp_err != nil {
				t.Errorf("HashedPassword does not match")
			}
		})
	}
}
