package snippet

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
)

type JSONPbWrapper struct {
	runtime.JSONPb
}

func (w *JSONPbWrapper) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		buffer, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("jsonpb wrapper io readall err: %+v, body: %+v", err, string(buffer))
		}
		err = w.JSONPb.Unmarshal(buffer, v)
		if err != nil {
			return fmt.Errorf("jsonpb wrapper unmarshal err: %+v, body: %+v", err, string(buffer))
		}
		return nil
	})
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	//clientDeadline := time.Now().Add(time.Duration(10) * time.Second) //grpc: the client connection is closing
	//ctx, cancel := context.WithDeadline(ctx, clientDeadline)
	defer cancel()

	marshalJsonPb := &JSONPbWrapper{JSONPb: runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{UseProtoNames: true}}}
	mux := runtime.NewServeMux(
		//runtime.WithRoutingErrorHandler(handleRoutingError),
		runtime.WithErrorHandler(CustomHTTPError),
		runtime.WithMarshalerOption(marshalJsonPb.ContentType("placeholder"), marshalJsonPb),
	)
