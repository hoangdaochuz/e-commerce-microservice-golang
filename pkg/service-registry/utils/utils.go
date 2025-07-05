package utils

import (
	"context"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// func ConvertMethodNameToSnakeCase(methodName string) string {
// 	var result strings.Builder
// 	for i, r := range methodName {
// 		if i > 0 && r >= 'A' && r <= 'Z' {
// 			result.WriteRune('_')
// 		}
// 		result.WriteRune(r)
// 	}
// 	return strings.ToLower(result.String())
// }

func implementsProtoMessage(message reflect.Type) bool {
	protoType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	return message.Implements(protoType)
}

func IsValidServiceMethod(method reflect.Method) bool {
	// check syntax signature: func (s service) NameFunc(ctx *context.Context, req proto.Message) (proto.Message, error)
	methodType := method.Type
	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return false
	}

	// check context type (context.Context)
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(1).Implements(ctxType) {
		return false
	}
	// check proto message type
	reqInputType := methodType.In(2)
	responseType := methodType.Out(0)
	if !implementsProtoMessage(reqInputType) || !implementsProtoMessage(responseType) {
		return false
	}
	// check type error return
	errType := reflect.TypeOf((*error)(nil)).Elem()
	return methodType.Out(1).Implements(errType)
}
