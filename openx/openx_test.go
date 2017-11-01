package openx

import (
	"os"
	"os/user"
	"testing"
)

// TestBadAuth ensures that an error is thrown when bad credentials are passed
func TestBadAuth(t *testing.T) {
	_, err := NewClient(Credentials{Domain: "domain", Realm: "realm", ConsumerKey: "key", ConsumerSecrect: "secret", Email: "email@gmail.com", Password: "password"}, false)
	if err == nil {
		t.Fatal("Calling new client should fail...it didn't")
	}
}

// TestParameters NewClient should fail if any of the parameters are empty
func TestParameters(t *testing.T) {
	var cc = []struct {
		Name string
		C    Credentials
	}{
		{"No Domain", Credentials{Domain: "", Realm: "realm", ConsumerKey: "key", ConsumerSecrect: "secret", Email: "email@gmail.com", Password: "password"}},
		{"No Realm", Credentials{Domain: "domain", Realm: "", ConsumerKey: "key", ConsumerSecrect: "secret", Email: "email@gmail.com", Password: "password"}},
		{"No Key", Credentials{Domain: "domain", Realm: "realm", ConsumerKey: "", ConsumerSecrect: "secret", Email: "email@gmail.com", Password: "password"}},
		{"No secret", Credentials{Domain: "domain", Realm: "realm", ConsumerKey: "key", ConsumerSecrect: "", Email: "email@gmail.com", Password: "password"}},
		{"No Email", Credentials{Domain: "domain", Realm: "realm", ConsumerKey: "key", ConsumerSecrect: "secret", Email: "", Password: "password"}},
		{"No Password", Credentials{Domain: "domain", Realm: "realm", ConsumerKey: "key", ConsumerSecrect: "secret", Email: "email@gmail.com", Password: ""}},
	}

	for _, c := range cc {
		t.Run(c.Name, func(t *testing.T) {
			_, err := NewClient(c.C, false)
			if err == nil {
				t.Fatalf("Test Name: %s, Message: Error should not be empty", c.Name)
			}
		})
	}
}

func TestConfigFileCreation(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Could not get home path:\n%v", err)
	}
	path := CreateConfigFileTemplate(usr.HomeDir)
	// remove the file once this this function completes
	defer os.Remove(path)
	t.Logf("TestConfigFileCreation File was removed: %s\n", path)
}

// TestBadAuthFromFile should fail because the JSON template is being used
func TestBadAuthFromFile(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Could not get home path:\n%v", err)
	}
	path := CreateConfigFileTemplate(usr.HomeDir)
	defer os.Remove(path)
	_, err = NewClientFromFile(path, false)
	if err == nil {
		t.Fatal("Calling new client should fail because the file json doesn't have the correct information")
	}
	t.Logf("TestBadAuthFromFile File was removed: %s\n", path)

}
