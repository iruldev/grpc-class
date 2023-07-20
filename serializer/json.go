package serializer

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
)

func ProtobufToJSON(message proto.Message) (string, error) {
	marshal, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}
