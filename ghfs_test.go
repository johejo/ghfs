package ghfs_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"testing/fstest"
	"testing/iotest"

	"golang.org/x/oauth2"

	"github.com/johejo/ghfs"
)

func newClient(t *testing.T) *http.Client {
	t.Helper()
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		t.Skip("no GITHUB_TOKEN")
	}
	return oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken}))
}

func TestFS(t *testing.T) {
	fsys := ghfs.New(newClient(t), "golang", "time")
	if err := fstest.TestFS(fsys, "README.md", "LICENSE", "rate/rate.go"); err != nil {
		t.Fatal(err)
	}
}

func TestIO(t *testing.T) {
	fsys := ghfs.New(newClient(t), "golang", "time")
	f, err := fsys.Open("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if err := iotest.TestReader(f, []byte("module golang.org/x/time\n")); err != nil {
		t.Fatal(err)
	}
}
