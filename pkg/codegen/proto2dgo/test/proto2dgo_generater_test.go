package proto2dgo_test

import (
	"testing"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/proto2dgo"
	"github.com/stretchr/testify/require"
)

func TestGeneraterDGoFile(t *testing.T) {
	protoFilePath := "./data/order.proto"
	output := "./gen/order.d.go"
	generater := proto2dgo.NewProto2dgoGenerater()
	err := generater.GenerateProto2Dgo(protoFilePath, output)
	require.NoError(t, err)
}
