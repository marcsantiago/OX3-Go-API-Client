package openx

import (
	"net/http"
	"strings"
	"testing"
)

// TestBadAuth ensures that an error is thrown when bad credentials are passed
func TestBadAuth(t *testing.T) {
	_, err := Newclient("domain", "realm", "key", "secret", "email@gmail.com", "password", false)
	if err == nil {
		t.Fatal("Calling new client should fail...it didn't")
	}
}

// TestParameters NewClient should fail if any of the parameters are empty
func TestParameters(t *testing.T) {
	_, err := Newclient("", "realm", "key", "secret", "email@gmail.com", "password", false)
	if err == nil {
		t.Fatal("should fail domain was left empty")
	}
	_, err = Newclient("domain", "", "key", "secret", "email@gmail.com", "password", false)
	if err == nil {
		t.Fatal("should fail realm was left empty")
	}
	_, err = Newclient("domain", "realm", "", "secret", "email@gmail.com", "password", false)
	if err == nil {
		t.Fatal("should fail key was left empty")
	}
	_, err = Newclient("domain", "realm", "key", "", "email@gmail.com", "password", false)
	if err == nil {
		t.Fatal("should fail secret was left empty")
	}
	_, err = Newclient("domain", "realm", "key", "secret", "", "password", false)
	if err == nil {
		t.Fatal("should fail email was left empty")
	}
	_, err = Newclient("", "realm", "key", "secret", "email@gmail.com", "", false)
	if err == nil {
		t.Fatal("should fail password was left empty")
	}
}

// TestGoodAuth should pass with good crendentials, the reason why they have to be hardcoded here is that the testing package
// ignores os.Getenv, it can be piped in however see https://stackoverflow.com/questions/33471976/set-environment-variable-for-go-tests
func TestGoodAuth(t *testing.T) {
	domain := ""
	realm := ""
	key := ""
	secret := ""
	email := ""
	password := ""

	shouldTest := func(vars ...string) bool {
		for _, v := range vars {
			if strings.TrimSpace(v) == "" {
				return false
			}
		}
		return true
	}

	if shouldTest(domain, realm, key, secret, email, password) {
		client, err := Newclient(domain, realm, key, secret, email, password, false)
		if err != nil {
			t.Fatalf("Either the credentials passed in were incorrect, or the client failed:\n %v", err)
		}
		res, err := client.Get("/report/get_reportlist", nil)
		if err != nil {
			t.Fatalf("Failed to get list of reports with /report/get_reportlist \n %v", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Requests failed, was expecting: %d instead I got %d", res.StatusCode, http.StatusOK)
		}
	}
}
