# ghfs

[![ci](https://github.com/johejo/ghfs/workflows/ci/badge.svg?branch=main)](https://github.com/johejo/ghfs/actions?query=workflow%3Aci)
[![Go Reference](https://pkg.go.dev/badge/github.com/johejo/ghfs.svg)](https://pkg.go.dev/github.com/johejo/ghfs)
[![codecov](https://codecov.io/gh/johejo/ghfs/branch/main/graph/badge.svg)](https://codecov.io/gh/johejo/ghfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/johejo/ghfs)](https://goreportcard.com/report/github.com/johejo/ghfs)

Package ghfs wraps the github v3 rest api with io/fs.
Files in the repository can be read in the same way as local files.

## Example

```go
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
```

## License

MIT

## Author

Mitsuo Heijo
