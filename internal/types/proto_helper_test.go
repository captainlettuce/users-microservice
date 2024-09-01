package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func checkProtobufAllFieldsSet(pb proto.Message) error {

	b, err := protojson.Marshal(pb)
	if err != nil {
		return fmt.Errorf("could not marshal proto: %w", err)
	}

	var anyMap map[string]any
	err = json.Unmarshal(b, &anyMap)
	if err != nil {
		return fmt.Errorf("could not unmarshal proto to map[string]any: %w", err)
	}

	protoFields := pb.ProtoReflect().Descriptor().Fields()

	if len(anyMap) != protoFields.Len() {
		return errors.New("unset fields found")
	}

	return nil
}
