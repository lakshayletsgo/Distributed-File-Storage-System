package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"io"
	"log"
	"os"
)

// This will create the parent folder as this name
const DefaultFolderName = "lakshaynetwork"

// This will create a new path for every file it stores
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	// hash[:] This will convert the string into hash

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
	// return strings.Join(paths, "/")
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	FileName string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]

}

// This will print the file name and the path
// The filename is the path
// expectedPathName := "32655/032aa/5f074/70b9e/205a4/a195a/a414e/07d89"
// This is the path to the file so the file name is
// expectedOriginalPathKey := "32655032aa5f07470b9e205a4a195aa414e07d89"

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

type StoreOpts struct {
	// This is the folder name of the root
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DefaultFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("Deleted this [%v] from disk\n", pathKey.FileName)
	}()
	// fmt.Printf("This is the path of the file to be deleted %v\n", pathKey.FullPath())

	// if err := os.RemoveAll(pathKey.FullPath()); err != nil {
	// 	return err
	// }
	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())
	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
	// n, f, err := s.readStream(key)
	// if err != nil {
	// 	return n, nil, err
	// }
	// defer f.Close()

	// buf := new(bytes.Buffer)
	// _, err = io.Copy(buf, f)
	// return n, buf, err

	return s.readStream(id, key)
}

func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	file, err := os.Open(pathKeyWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil

}

func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)
	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		fmt.Printf("Inside if 2 : %v\n", err)

		return 0, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	n, err := copyDecrypt(encKey, r, f)

	return int64(n), err
}

// This is used to write/store the data
func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)
	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		return 0, err
	}
	// buf := new(bytes.Buffer)
	// io.Copy(buf, r)

	// filenameBytes := md5.Sum(buf.Bytes())
	// fileName := hex.EncodeToString(filenameBytes[:])

	// fileName := "somefilename"
	// This will be the file name of the created file

	// fullPath := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	return io.Copy(f, r)

	// return n, err
}
