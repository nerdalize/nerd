package provider

import (
	"io/ioutil"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

//Disk provides nerdalize credentials from a file on disk. The default file can be found in TokenFilename().
type Disk struct {
	*ProviderBasis
	DiskLocation string
}

func NewDiskCredentials() *credentials.NerdAPI {
	return credentials.NewNerdAPI(NewDisk())
}

func NewDisk() *Disk {
	return &Disk{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
		},
	}
}

func (e *Disk) Retrieve() (*credentials.NerdAPIValue, error) {
	if e.DiskLocation == "" {
		filename, err := TokenFilename()
		if err != nil {
			return nil, err
		}
		e.DiskLocation = filename
	}
	t, err := ioutil.ReadFile(e.DiskLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read nerd token at '%v'", e.DiskLocation)
	}
	token := string(t)
	if token == "" {
		return nil, errors.Errorf("file %v is empty", e.DiskLocation)
	}
	if e.TokenDecoder == nil {
		e.TokenDecoder = credentials.DecodeToken
	}
	claims, err := e.TokenDecoder(token)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", token)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return nil, errors.Wrapf(err, "nerd token '%v' is invalid", token)
	}
	return &credentials.NerdAPIValue{NerdToken: token}, nil
}
