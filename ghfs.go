// Package ghfs wraps the github v3 rest api with io/fs.
// Files in the repository can be read in the same way as local files.
package ghfs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/google/go-github/v33/github"
)

var (
	ctx = context.TODO()
)

// FS is a file system using github v3 reset api.
// ref: https://docs.github.com/en/rest/reference/repos#get-repository-content
type FS struct {
	owner string
	repo  string
	rs    *github.RepositoriesService
}

var (
	_ fs.FS         = (*FS)(nil)
	_ fs.ReadFileFS = (*FS)(nil)
	_ fs.ReadDirFS  = (*FS)(nil)
)

// New returns a new github file sytem.
func New(c *http.Client, owner, repo string) fs.FS {
	return &FS{owner: owner, repo: repo, rs: github.NewClient(c).Repositories}
}

// Open implementes fs.FS and opens a new file asd fs.File.
func (f *FS) Open(name string) (fs.File, error) {
	fc, dc, err := f.getContents(name)
	if err != nil {
		return nil, err
	}
	if dc != nil {
		files := make([]*repoContent, len(dc))
		for i := range dc {
			files[i] = &repoContent{c: dc[i]}
		}
		return &openDir{d: dir{name: name}, files: files}, nil
	}

	data, err := fc.GetContent()
	if err != nil {
		return nil, err
	}
	return &openFile{c: &repoContent{c: fc}, data: data}, nil
}

// ReadFile implementes fs.ReadFileFS.
func (f *FS) ReadFile(name string) ([]byte, error) {
	fc, _, err := f.getContents(name)
	if err != nil {
		return nil, err
	}
	if fc == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	c, err := fc.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(c), nil
}

// ReadFile implementes fs.ReadDirFS.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	_, dc, err := f.getContents(name)
	if err != nil {
		return nil, err
	}
	if dc == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	files := make([]fs.DirEntry, len(dc))
	for i := range dc {
		files[i] = &repoContent{c: dc[i]}
	}
	return files, nil
}

func (f *FS) getContents(name string) (fc *github.RepositoryContent, dc []*github.RepositoryContent, err error) {
	if !fs.ValidPath(name) {
		return nil, nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	fc, dc, _, err = f.rs.GetContents(ctx, f.owner, f.repo, name, nil)
	if err != nil {
		return nil, nil, err
	}
	return fc, dc, nil
}

type openFile struct {
	c      *repoContent
	data   string
	offset int64
}

var (
	_ fs.File     = (*openFile)(nil)
	_ io.Seeker   = (*openFile)(nil)
	_ io.ReaderAt = (*openFile)(nil)
)

func (f *openFile) Stat() (fs.FileInfo, error) { return f.c, nil }
func (f *openFile) Close() error               { return nil }

func (f *openFile) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	if f.offset < 0 {
		return 0, &fs.PathError{Op: "read", Path: f.c.Name(), Err: fs.ErrInvalid}
	}
	n := copy(b, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *openFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		// offset += 0
	case 1:
		offset += f.offset
	case 2:
		offset += int64(len(f.data))
	}
	if offset < 0 || offset > int64(len(f.data)) {
		return 0, &fs.PathError{Op: "seek", Path: f.c.Name(), Err: fs.ErrInvalid}
	}
	f.offset = offset
	return offset, nil
}

func (f *openFile) ReadAt(b []byte, offset int64) (int, error) {
	if offset < 0 || offset > int64(len(f.data)) {
		return 0, &fs.PathError{Op: "read", Path: f.c.Name(), Err: fs.ErrInvalid}
	}
	n := copy(b, f.data[offset:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

type openDir struct {
	d      dir
	files  []*repoContent
	offset int
}

var _ fs.File = (*openDir)(nil)
var _ fs.ReadDirFile = (*openDir)(nil)

func (d *openDir) Stat() (fs.FileInfo, error) { return d.d, nil }
func (d *openDir) Close() error               { return nil }

func (d *openDir) Read(b []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.d.Name(), Err: errors.New("is a directory")}
}

func (d *openDir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.files) - d.offset
	if count > 0 && n > count {
		n = count
	}
	if n == 0 {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	list := make([]fs.DirEntry, n)
	for i := range list {
		list[i] = d.files[d.offset+i]
	}
	d.offset += n
	return list, nil
}

type repoContent struct {
	c *github.RepositoryContent
}

var _ fs.FileInfo = (*repoContent)(nil)
var _ fs.DirEntry = (*repoContent)(nil)

func (c *repoContent) Type() fs.FileMode {
	return c.Mode().Type()
}

func (c *repoContent) Info() (fs.FileInfo, error) {
	return c, nil
}

func (c *repoContent) ModTime() time.Time {
	return time.Time{}
}

func (c *repoContent) Sys() interface{} {
	return c.c
}

func (c *repoContent) Name() string {
	return c.c.GetName()
}

func (c *repoContent) Size() int64 {
	return int64(c.c.GetSize())
}

func (c *repoContent) Mode() fs.FileMode {
	switch c.c.GetType() {
	case "dir":
		return fs.ModeDir | 0o555
	case "file":
		return 0o444
	default:
		return 0
	}
}

func (c *repoContent) IsDir() bool {
	return c.c.GetType() == "dir"
}

type dir struct {
	name string
}

var _ fs.FileInfo = dir{}

func (d dir) Name() string {
	return d.name
}

func (d dir) Size() int64 {
	return 0
}

func (d dir) Mode() fs.FileMode {
	return fs.ModeDir | 0o555
}

func (d dir) ModTime() time.Time {
	return time.Time{}
}

func (d dir) IsDir() bool {
	return true
}

func (d dir) Sys() interface{} {
	return d.name
}
