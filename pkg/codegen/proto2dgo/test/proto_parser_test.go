package proto2dgo_test

import (
	"testing"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/proto2dgo"
	"github.com/stretchr/testify/require"
)

var filepath = "./data/order.proto"
var protoParser = proto2dgo.NewProtoParser()

// Test get syntax proto file
func TestGetHeaderOfProto(t *testing.T) {
	protoModel, err := protoParser.ParseProtoFile(filepath)
	require.NoError(t, err)
	require.NotNil(t, protoModel)
	t.Run("Test_GetSyntaxProtofile", func(t *testing.T) {
		require.Equal(t, protoModel.Syntax, "proto3")
	})
	t.Run("Test_GetProtoPackage", func(t *testing.T) {
		require.Equal(t, protoModel.ProtoPackage, "order")
	})
	t.Run("Test_GetGoPackage", func(t *testing.T) {
		require.Equal(t, protoModel.GoPackage, "order/api/order")
	})
	t.Run("Test_GetImportPath", func(t *testing.T) {
		require.Equal(t, len(protoModel.ImportPaths), 2)
		require.Equal(t, protoModel.ImportPaths[0].Path, "example/abc.proto")
		require.Equal(t, protoModel.ImportPaths[1].Mode, proto2dgo.ImportMode("public"))
		require.Equal(t, protoModel.ImportPaths[1].Path, "example/test.proto")
	})
	t.Run("Test_GetEnum", func(t *testing.T) {
		require.Equal(t, len(protoModel.Enums), 1)
		require.Equal(t, protoModel.Enums[0].Name, "OrderStatus")
		require.Equal(t, len(protoModel.Enums[0].EnumFields), 3)
		require.Equal(t, protoModel.Enums[0].EnumFields[0].Key, "FAIL")
		require.Equal(t, protoModel.Enums[0].EnumFields[0].Value, 0)
		require.Equal(t, protoModel.Enums[0].EnumFields[1].Key, "PENDING")
		require.Equal(t, protoModel.Enums[0].EnumFields[1].Value, 1)
		require.Equal(t, protoModel.Enums[0].EnumFields[2].Key, "SUCCESS")
		require.Equal(t, protoModel.Enums[0].EnumFields[2].Value, 2)
	})

	t.Run("Test_GetService", func(t *testing.T) {
		require.Equal(t, len(protoModel.Services), 1)
		require.Equal(t, protoModel.Services[0].Name, "OrderService")
		require.Equal(t, len(protoModel.Services[0].Methods), 2)

		require.Equal(t, protoModel.Services[0].Methods[0].Name, "CreateOrder")
		require.Equal(t, protoModel.Services[0].Methods[0].ConstantName, "ORDER_CREATE_ORDER")
		require.Equal(t, protoModel.Services[0].Methods[0].RequestType, "CreateOrderRequest")
		require.Equal(t, protoModel.Services[0].Methods[0].ResponseType, "CreateOrderResponse")
	})

	t.Run("Test_GetMessage", func(t *testing.T) {
		require.Equal(t, len(protoModel.Messages), 5)
		require.Equal(t, protoModel.Messages[0].MessageName, "GetOrderByIdRequest")
		require.Equal(t, len(protoModel.Messages[0].Fields), 1)
		require.Equal(t, protoModel.Messages[0].Fields[0].Name, "Id")
		require.Equal(t, protoModel.Messages[0].Fields[0].Type, "string")
		require.Equal(t, protoModel.Messages[0].Fields[0].Order, 1)
		require.Equal(t, protoModel.Messages[0].Fields[0].IsOptional, false)
		require.Equal(t, protoModel.Messages[0].Fields[0].IsRepeat, false)

		require.Equal(t, protoModel.Messages[4].Fields[0].IsRepeat, true)

		require.Equal(t, protoModel.Messages[4].MessageName, "TestMessage")
		require.Equal(t, len(protoModel.Messages[4].Fields), 5)
		require.Equal(t, protoModel.Messages[4].Fields[3].Name, "importFields")
		require.Equal(t, protoModel.Messages[4].Fields[3].Type, "test.TestItem")
		require.Equal(t, protoModel.Messages[4].Fields[3].Order, 4)
		require.Equal(t, protoModel.Messages[4].Fields[3].IsOptional, false)
		require.Equal(t, protoModel.Messages[4].Fields[3].IsRepeat, false)

	})
}
