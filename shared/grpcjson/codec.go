package grpcjson

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// Name is the gRPC "content-subtype" for this codec. ok
const Name = "json"

type codec struct{}

func (codec) Name() string { return Name }
func (codec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func init() {
	encoding.RegisterCodec(codec{})
}
