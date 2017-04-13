package tpr

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fakeCodec struct {
	encoded      runtime.Object
	encodedBytes []byte
	encodeErr    error

	decodedBytes []byte
	decodeInto   runtime.Object
	decodeErr    error
}

func (f *fakeCodec) Encode(obj runtime.Object, w io.Writer) error {
	if _, err := w.Write(f.encodedBytes); err != nil {
		return err
	}
	f.encoded = obj
	return f.encodeErr
}

func (f *fakeCodec) Decode(
	data []byte,
	defaults *schema.GroupVersionKind,
	into runtime.Object,
) (runtime.Object, *schema.GroupVersionKind, error) {
	f.decodedBytes = data
	if f.decodeErr != nil {
		return nil, nil, f.decodeErr
	}
	return f.decodeInto, nil, nil
}
