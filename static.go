// Code generated by "esc -o static.go static/"; DO NOT EDIT.

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/static/gone.tmpl": {
		local:   "static/gone.tmpl",
		size:    2627,
		modtime: 1506407493,
		compressed: `
H4sIAAAAAAAC/5RWUY/bNgx+dn6F5kPrHHCxk1vRFa7jAbsWQ4FhLbp7WYtiUCzaVk+WPIlJLjH83wdJ
jpMc0nV7ikORH8mPn2l1HYOSSyAhchQQ9v2vSgK55w2Qe02LB9Ck63hJ4k/7fd93HULTCopAQiOV2tuI
rgPJxp/J5Ig5unza75/XIARvX19yw51LPQkyweUDqTWUy6hGbNMkKZVEE1dKVQJoy01cqCYpjPm5pA0X
u+VHtVKo0h/n8+dmvTKAS0GRy5tip7kQvIiIBrGMXA5TA2BEcNfCMkJ4RAsU5ZMgc8f5JAhWiu1INwmC
wCae+SQpiXya6IYYKs3MgObla+vVUF1xmZLbZ+TlM2vpJ0FADWcwoAhFMSWaVzUejq9W6tGfrpRmoFOC
NZfEKMEZuXpx+4q+uHXgK1o8VFqtJUvJFZRlCT5pSxnjskrJApqTKmYCShyNNlOsoVCamVRiPStqLtgU
NiCvh+zn8DDAnwYSZCext0PgljOsU7KYz8eWY1RIxWlXM1TtWWcrQYuHkSLvaocwYzYXRa5kSqSS4Foq
lFA6JVcvi58WBRzC1kMKwQ3O3NBmdprHOF+8ZfskARW8kidDGBu4hcdDEDIfMHI7J4vxNEsGgVxQb1FT
jV69ptC8Ra+v0OnrK91Qbw2J0cUytKo2aZJst9tB1E7QXw1teZhniXfOvwdmpTqEC0XZNNpws6aC7x2N
0Q2JFvE8uiFd1NLigVZgovRzVCgNrtzoS3/9+ghhAN/L3xRld1QIq4op03R7Zz2dW7mWhcUlo3k6KGFD
NWEUKVkSCVsy4J0VE7+hSO/pSsDUgQXWP6aM3SmxbuQ0Mqi5rGzNd4IaE132kutmBdp62eX0xOmj2prp
Z2vpOk1lBSR2WGD6/nPUdf5f31tGuvgD6AIk9v2Xm8M8gyD44iFtR46jf2vpAwdPA1PFugGJcQX4VoB9
/GX3jk0jB/EX45vo+oirWhttyNKTFwiowL58kVVvZE29174Nji3ZU9vhzSHw+ijHQSYX9OiWj9eje7RS
yQz4AXK2DFfq0eknQzuU/DJr1pqhdqdBhizPqN/LYXKkM8yPz1lC8yxBNkaQwh4sQ/fWOc8/Wsf66JUl
Q4KTKXwjZ5jf2/3yzSR++xxTOe9Lqc7wL37OvtPEOybgAnCWjGxmycB2PnlKvd8VzonxzdFkhWLffsY3
TxCCLBmmeGn1KIkgh+VzzD+2WuduOFmC9Wj5nTZwZnBTOVgO7YyC+Oi/A244GeqRD28OB5T/KQ/PfWxL
OaXyP2jmWN+gmJH3C/RopRw32Q9v3t/d//nhLamxEfkk8z9BVgP1qO7ycyYHfx0isUvtTq1fA0jddrD3
jHCN5exV6MvZcqwtW6UGU/e997TLfgZ/r/nGMuaOQjIMbRna7qBQkg1UHd6BM1WOV6Mzs9eRrc4y4NuY
BJm9vLg668U3m6kX+RMwvzA82HmWQV6HPB4+Szx9h4L/CQAA///o0wxaQwoAAA==
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/static": {
		isDir: true,
		local: "static",
	},
}