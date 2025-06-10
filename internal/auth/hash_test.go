package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestJWT(t *testing.T) {
	dur, _ := time.ParseDuration("15m")
	testID, _ := uuid.Parse("518e1d75-cf97-4b27-b6ff-5c70a47fa66c")
	cases := []struct {
		testUserID      uuid.UUID
		testTokenSecret string
		testExpiresIn   time.Duration
	}{
		{
			testUserID:      testID,
			testTokenSecret: "wubbawubba",
			testExpiresIn:   dur,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			res, err := MakeJWT(c.testUserID, c.testTokenSecret, c.testExpiresIn)
			if err != nil {
				t.Errorf("MakeJWT broke")
			}
			ret_id, err := ValidateJWT(res, c.testTokenSecret)
			if err != nil {
				t.Errorf("ValidateJWT broke: %s", err)
			}
			if ret_id != c.testUserID {
				t.Errorf("Expected: %s; got %s", c.testUserID, ret_id)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		key http.Header
		val string
	}{
		{
			key: http.Header{
				"Content-Type":  {"application/json"},
				"Authorization": {"Bearer justsometoken"},
			},
			val: "justsometoken",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			res, err := GetBearerToken(c.key)
			if err != nil {
				t.Errorf("GetBearerToken Broke")
			}
			if c.val != res {
				t.Errorf("Expected %s got %s instead", c.val, res)
			}

		})
	}
}
