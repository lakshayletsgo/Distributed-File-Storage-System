package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	// "fmt"

	"testing"
)

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}
func tearDown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}

func TestPathTransformFunc(t *testing.T) {
	key := "mynameislakshay"
	pathKey := CASPathTransformFunc(key)
	// fmt.Println(pathname)
	expectedPathName := "32655/032aa/5f074/70b9e/205a4/a195a/a414e/07d89"
	expectedOriginalPathKey := "32655032aa5f07470b9e205a4a195aa414e07d89"
	if pathKey.PathName != expectedPathName {
		t.Errorf("We have this pathname %v instead of this %v", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedOriginalPathKey {
		t.Errorf("We have this pathname %v instead of this %v", pathKey.FileName, expectedOriginalPathKey)
	}
}

func TestStore(t *testing.T) {

	s := newStore()
	id := generateId()
	defer tearDown(t, s)
	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("Lakshay_%d", i)
		data := []byte("My JPG Files")
		// This will be the data written in the file

		// Special Pictures will be the folder name under which the file name somefilename will be there
		if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if !s.Has(id, key) {
			t.Errorf("Expected to have this %s key but didn't find it", key)
		}

		_, r, err := s.Read(id, key)
		if err != nil {
			t.Errorf("This is the error while reading %v\n", err)
		}

		b, _ := ioutil.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("Entered this data %v got thsi %v", data, b)
		}

		if err := s.Delete(id, key); err != nil {
			t.Errorf("There is error while deleting the file %s", err)
		}
		if s.Has(id, key) {
			t.Errorf("Expected to not have this %s key", key)
		}
	}
}
