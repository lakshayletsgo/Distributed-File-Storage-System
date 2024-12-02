package main

import (
	"bytes"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	payload := "Who are you"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := EncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Errorf("Error in copying the data %v\n", err)
	}

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Errorf("There is an error while decrypting the file data %v\n", err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}
	if out.String() != payload {
		t.Errorf("The output is not the same as the input\n")
	}

}
