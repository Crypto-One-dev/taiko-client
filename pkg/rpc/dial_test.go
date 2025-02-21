package rpc

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/stretchr/testify/require"
	"github.com/taikoxyz/taiko-client/pkg/jwt"
)

func TestDialEngineClientWithBackoff(t *testing.T) {
	jwtSecret, err := jwt.ParseSecretFromFile(os.Getenv("JWT_SECRET"))

	require.Nil(t, err)
	require.NotEmpty(t, jwtSecret)

	client, err := DialEngineClientWithBackoff(
		context.Background(),
		os.Getenv("L2_NODE_ENGINE_ENDPOINT"),
		string(jwtSecret),
	)

	require.Nil(t, err)

	var result beacon.ExecutableDataV1
	err = client.CallContext(context.Background(), &result, "engine_getPayloadV1", beacon.PayloadID{})

	require.Equal(t, beacon.UnknownPayload.Error(), err.Error())
}

func TestDialClientWithBackoff(t *testing.T) {
	client, err := DialClientWithBackoff(context.Background(), os.Getenv("L2_NODE_ENDPOINT"))
	require.Nil(t, err)

	genesis, err := client.HeaderByNumber(context.Background(), common.Big0)
	require.Nil(t, err)

	require.Equal(t, common.Big0.Uint64(), genesis.Number.Uint64())
}
