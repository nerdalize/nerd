package conf

import (
	"io/ioutil"
	"testing"
)

const test_conf = `
{
  "auth": {
      "public_key": "test_key",
      "api_endpoint": "test_url"
  }
}
`

func TestFromFile(t *testing.T) {
	temp, err := ioutil.TempFile("/tmp", "nerd_conf")
	if err != nil {
		t.Fatalf("Unexpected error for temp file: %v", err)
	}
	temp.WriteString(test_conf)
	SetLocation(temp.Name())
	conf, err := Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	auth := conf.Auth
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if auth.APIEndpoint != "test_url" {
		t.Errorf("Expected api_endpoint %v but got %v", "test_url", auth.APIEndpoint)
	}
	if auth.PublicKey != "test_key" {
		t.Errorf("Expected api_endpoint %v but got %v", "test_key", auth.PublicKey)
	}
}
