package ghfs_test

import (
	"context"
	"io"
	"os"

	"github.com/johejo/ghfs"
	"golang.org/x/oauth2"
)

func Example() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	c := oauth2.NewClient(ctx, ts)

	fsys := ghfs.New(c, "golang", "go")
	f, err := fsys.Open("README.md")
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, f)
}
