package v1batch

import (
	"fmt"
	"net/url"
	"testing"
)

type logger struct{}

func (l *logger) Error(args ...interface{}) {
	fmt.Println(args)
}

func (l *logger) Debugf(a string, args ...interface{}) {
	fmt.Printf(a, args)
}

func TestDataset(t *testing.T) {
	base, err := url.Parse("https://batch.nerdalize.com/uni031-boris")
	if err != nil {
		panic(err)
	}
	_ = NewClient(ClientConfig{
		Base:        base,
		Logger:      &logger{},
		JWTProvider: NewStaticJWTProvider("abcd"),
	})
	// cl.CreateDataset("6de308f4-face-11e6-bc64-92361f002671")
}
