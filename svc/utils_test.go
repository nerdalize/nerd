package svc_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/svc"
)

func isNilErr(err error) bool {
	return err == nil
}

type testingDI struct {
	kube kubernetes.Interface
}

func (di *testingDI) Kube() kubernetes.Interface {
	return di.kube
}

func testDI(tb testing.TB) svc.DI {
	tb.Helper()

	hdir, err := homedir.Dir()
	ok(tb, err)

	tdi := &testingDI{}
	kcfg, err := clientcmd.BuildConfigFromFlags("", filepath.Join(hdir, ".kube", "config"))
	ok(tb, err)

	if !strings.Contains(fmt.Sprintf("%#v", kcfg), "minikube") {
		tb.Skipf("kube config needs to contain 'minikube' for local testing")
	}

	tdi.kube, err = kubernetes.NewForConfig(kcfg)
	ok(tb, err)

	return tdi
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
