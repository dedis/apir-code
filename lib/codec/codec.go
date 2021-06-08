package codec

import (
	"encoding/json"

	"github.com/si-co/vpir-code/lib/proto"
	"golang.org/x/xerrors"
)

// Codec implements a custom gRPC codec, which in our case is faster and more
// memory efficient.
type Codec struct{}

// Marshal implements encoding.Codec
func (cb *Codec) Marshal(v interface{}) ([]byte, error) {
	switch msg := v.(type) {
	case *proto.DatabaseInfoRequest:
		return json.Marshal(v)
	case *proto.DatabaseInfoResponse:
		return json.Marshal(v)
	case *proto.QueryRequest:
		return msg.Query, nil
	case *proto.QueryResponse:
		return msg.Answer, nil
	default:
		return nil, xerrors.Errorf("unknown message %T", msg)
	}
}

// Unmarshal implements encoding.Codec
func (cb *Codec) Unmarshal(data []byte, v interface{}) error {
	switch msg := v.(type) {
	case *proto.DatabaseInfoRequest:
		return json.Unmarshal(data, v)
	case *proto.DatabaseInfoResponse:
		return json.Unmarshal(data, v)
	case *proto.QueryRequest:
		msg.Query = data
		return nil
	case *proto.QueryResponse:
		msg.Answer = data
		return nil
	default:
		return xerrors.Errorf("unknown message %T", msg)
	}
}

// Name implements encoding.Codec
func (cb *Codec) String() string {
	return "worker.PayloadCodec"
}

// Name implements encoding.Codec
func (cb *Codec) Name() string {
	return "worker.PayloadCodec"
}
