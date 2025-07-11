package trcshfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/trimble-oss/tierceron-core/v2/core/coreconfig"
	"github.com/trimble-oss/tierceron-core/v2/trcshfs/trcshio"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

type TrcshMemFs struct {
	BillyFs      *billy.Filesystem
	MemCacheLock sync.Mutex
}

func NewTrcshMemFs() *TrcshMemFs {
	billyFs := memfs.New()
	return &TrcshMemFs{
		BillyFs: &billyFs,
	}
}

func (t *TrcshMemFs) WriteToMemFile(coreConfig *coreconfig.CoreConfig, byteData *[]byte, path string) {

	t.MemCacheLock.Lock()
	if _, err := (*t.BillyFs).Stat(path); errors.Is(err, os.ErrNotExist) {
		if strings.HasPrefix(path, "./") {
			path = strings.TrimLeft(path, "./")
		}
		memFile, err := t.Create(path)
		if err != nil {
			coreConfig.Log.Printf("Error creating memfile %s: %v", path, err)
		}
		memFile.Write(*byteData)
		memFile.Close()
		t.MemCacheLock.Unlock()
		coreConfig.Log.Printf("Wrote memfile: %s", path)
	} else {
		t.MemCacheLock.Unlock()
		coreConfig.Log.Printf("Memfile already exists: %s", path)
	}
}

func (t *TrcshMemFs) ReadDir(path string) ([]os.FileInfo, error) {
	t.MemCacheLock.Lock()
	defer t.MemCacheLock.Unlock()
	return (*t.BillyFs).ReadDir(path)
}

func (t *TrcshMemFs) WalkCache(path string, nodeProcessFunc func(string) error) {
	t.MemCacheLock.Lock()
	defer t.MemCacheLock.Unlock()
	filestack := []string{path}
	var p string

summitatem:

	if len(filestack) == 0 {
		return
	}
	p, filestack = filestack[len(filestack)-1], filestack[:len(filestack)-1]

	if fileset, err := (*t.BillyFs).ReadDir(p); err == nil {
		if path != "." && len(fileset) > 0 {
			filestack = append(filestack, p)
		}
		for _, file := range fileset {
			if file.IsDir() {
				filestack = append(filestack, fmt.Sprintf("%s/%s", p, file.Name()))
			} else {
				nodeProcessFunc(fmt.Sprintf("%s/%s", p, file.Name()))
			}
		}
	}
	nodeProcessFunc(p)

	goto summitatem
}

func (t *TrcshMemFs) ClearCache(path string) {
	if path == "." {
		t.MemCacheLock.Lock()
		defer t.MemCacheLock.Unlock()
		t.BillyFs = nil
		newBilly := memfs.New()
		t.BillyFs = &newBilly
	} else {
		t.WalkCache(path, t.Remove)
	}
}

func (t *TrcshMemFs) SerializeToMap(path string, configCache map[string]any) {
	t.WalkCache(path, func(path string) error {
		if fileReader, err := t.Open(path); err == nil {
			bytesBuffer := new(bytes.Buffer)

			io.Copy(bytesBuffer, fileReader)
			configCache[path] = bytesBuffer.Bytes()
		}
		return nil
	})
}

func (t *TrcshMemFs) Create(filename string) (trcshio.TrcshReadWriteCloser, error) {
	return (*t.BillyFs).Create(filename)
}

func (t *TrcshMemFs) Open(filename string) (trcshio.TrcshReadWriteCloser, error) {
	return (*t.BillyFs).Open(filename)
}

func (t *TrcshMemFs) Stat(filename string) (os.FileInfo, error) {
	return (*t.BillyFs).Stat(filename)
}

func (t *TrcshMemFs) Remove(filename string) error {
	if filename != "." {
		return (*t.BillyFs).Remove(filename)
	} else {
		return nil
	}
}

func (t *TrcshMemFs) Lstat(filename string) (os.FileInfo, error) {
	return (*t.BillyFs).Lstat(filename)
}

func (t *TrcshMemFs) Walk(root string, walkFn func(path string, isDir bool) error) error {
	infos, err := (*t).ReadDir(root)
	if err != nil {
		return err
	}
	for _, info := range infos {
		fullPath := path.Join(root, info.Name())
		err := walkFn(fullPath, info.IsDir())
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = t.Walk(fullPath, walkFn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
