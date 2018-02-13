package transfer_test

import (
	"testing"

	transfer "github.com/nerdalize/nerd/pkg/transfer/v2"
)

func TestKubeManagerCreate(t *testing.T) {
	var (
		mgr transfer.Manager
		err error
	)

	//@TODO setup test visor
	mgr, err = transfer.NewKubeManager(nil)
	if err != nil {
		t.Fatal(err)
	}

	_ = mgr

}
