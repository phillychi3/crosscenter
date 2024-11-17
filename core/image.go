package core

import (
	"bytes"
	"io"
	"net/http"
)

func GetImageBytes(image string) ([]byte, error) {
	resp, err := http.Get(image)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
