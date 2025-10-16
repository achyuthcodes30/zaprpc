package zaprpc

import (
	"encoding/gob"
	"io"
)

type Codec interface {
	Name() string
	Marshal(w io.Writer, v any) error
	Unmarshal(r io.Reader, v any) error
}

type GOBCodec struct{}

func (g *GOBCodec) Name() string {
	return "gob"
}

func (g *GOBCodec) Marshal(w io.Writer, v any) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(v)
}

func (g *GOBCodec) Unmarshal(r io.Reader, v any) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(v)
}
