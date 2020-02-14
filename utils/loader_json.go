package utils

import (
	"encoding/json"
	"io"
)

func LoadJSON(r io.Reader, newElement func() interface{}, callback func(interface{}) error) error {
	dec := json.NewDecoder(r)
	for {
		el := newElement()
		if err := dec.Decode(&el); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := callback(el); err != nil {
			return err
		}
	}
	return nil
}
