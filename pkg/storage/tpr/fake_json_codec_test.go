package tpr

import (
	"bytes"
	"encoding/json"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fakeJSONCodec is a codec that simply json encodes and decodes objects to encode and decode
type fakeJSONCodec struct {
	encodeErr error
	decodeErr error
}

func (f *fakeJSONCodec) Encode(obj runtime.Object, w io.Writer) error {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return err
	}
	return f.encodeErr
}

func (f *fakeJSONCodec) Decode(
	data []byte,
	defaults *schema.GroupVersionKind,
	into runtime.Object,
) (runtime.Object, *schema.GroupVersionKind, error) {
	if f.decodeErr != nil {
		return nil, nil, f.decodeErr
	}
	b := bytes.NewBuffer(data)
	if err := json.NewDecoder(b).Decode(into); err != nil {
		return nil, nil, err
	}
	return into, nil, nil
}
