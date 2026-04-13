package siwk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMessage(t *testing.T) {
	negativeExamples := []struct {
		example string
		error   error
	}{
		{
			example: "",
			error:   ErrMessageTooShort,
		},
		{
			example: "\n\n\n\n",
			error:   ErrMessageTooShort,
		},
		{
			example: "domain.com whatever\n\n\n\n\n\n",
			error:   ErrInvalidHeader,
		},
		{
			example: "******* wants you to sign in with your Kaspa address:\n\n\n\n\n\n",
			error:   ErrInvalidDomain,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\n***************************************\n\n\n\n\n",
			error:   ErrInvalidAddress,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\nURI: https://google.com\n\n\n",
			error:   ErrThirdLineNotEmpty,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nNot Parsable\n",
			error:   errUnparsableLine(5),
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: ***\nIssued At: 2025-01-01T00:00:00Z",
			error:   ErrInvalidURI,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://google.com\nIssued At: not-a-timestamp",
			error:   ErrInvalidIssuedAt,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://google.com\nIssued At: 2025-01-01T00:00:00Z\nExpiration Time: not-a-timestamp",
			error:   ErrInvalidExpirationTime,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://google.com\nIssued At: 2025-01-01T00:00:00Z\nNot Before: not-a-timestamp",
			error:   ErrInvalidNotBefore,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 2\nIssued At: 2025-01-01T00:00:00Z\nURI: https://google.com\n",
			error:   errUnsupportedVersion("2"),
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nNonce: 12345678\nIssued At: 2025-01-01T00:00:00Z\n\n",
			error:   ErrMissingURI,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nNonce: 12345678\nResources:\n- https://google.com\n",
			error:   ErrMissingIssuedAt,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nNonce: 12345678\nIssued At: 2025-01-02T00:00:00Z\nExpiration Time: 2025-01-01T00:00:00Z\n",
			error:   ErrIssuedAfterExpiration,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nNonce: 12345678\nIssued At: 2025-01-01T00:00:00Z\nExpiration Time: 2025-01-02T00:00:00Z\nNot Before: 2025-01-03T00:00:00Z\n",
			error:   ErrNotBeforeAfterExpiration,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nIssued At: 2025-01-01T00:00:00Z\nResources:\n- https://google.com\n- ***\n",
			error:   errInvalidResource(1),
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nVersion: 1\nURI: https://domain.com\nIssued At: 2025-01-01T00:00:00Z\nChain ID: random:mainnet",
			error:   ErrInvalidNetworkID,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nNonce: short\nIssued At: 2025-01-01T00:00:00Z\n",
			error:   ErrInvalidNonce,
		},
		{
			example: "domain.com wants you to sign in with your Kaspa address:\nkaspa:qqk948c2dy6cp0vdg7fqx9xttc47q4qdazunhmfv8u24v77uvmxhycc2uj3yn\n\nStatement\n\nVersion: 1\nURI: https://domain.com\nIssued At: 2025-01-01T00:00:00Z\n",
			error:   ErrMissingNonce,
		},
	}

	for i, example := range negativeExamples {
		_, err := ParseMessage(example.example)

		t.Run(fmt.Sprintf("negative example %d", i), func(t *testing.T) {
			require.NotNil(t, err)
			require.Equal(t, example.error.Error(), err.Error())
		})
	}

	positiveExamples := []struct {
		message   string
		signature string
	}{
		{
			message:   "example.com wants you to sign in with your Kaspa address:\nkaspa:qpvmk3kpwuyavpgh0gcsmayrecsxlxuhy0qjadzh4v98umk9fs2c69z4q90j7\n\nSign in to Example App\n\nURI: https://example.com\nVersion: 1\nChain ID: kaspa_mainnet\nNonce: 12345678\nIssued At: 2025-11-03T10:55:43.367Z",
			signature: "73fc16d604941add11c0601f95b5d6ac3c297340dcd31dd1932c426ca9150e083c28aeaefdc361256da50e78d6b551eb175e82a000aa8c90f95c64d8a10c3863",
		},
		{
			message:   "example.com wants you to sign in with your Kaspa address:\nkaspa:qpvmk3kpwuyavpgh0gcsmayrecsxlxuhy0qjadzh4v98umk9fs2c69z4q90j7\n\nURI: https://example.com\nVersion: 1\nChain ID: kaspa_mainnet\nNonce: 12345678\nIssued At: 2025-11-03T10:55:43.367Z",
			signature: "e816f93857b356eb526fb2c3b5d7ecafa9e7c7b2e0014c2ff2a70647cebca0086fab6df228d71fae975f72d9b592f89a7f1460dd0106bfa08e5637b12c4a9768",
		},
	}

	for i, example := range positiveExamples {
		t.Run(fmt.Sprintf("positive example %d", i), func(t *testing.T) {
			parsed, err := ParseMessage(example.message)

			require.Nil(t, err)
			require.Equal(t, "example.com", parsed.Domain)
			require.Equal(t, "kaspa:qpvmk3kpwuyavpgh0gcsmayrecsxlxuhy0qjadzh4v98umk9fs2c69z4q90j7", parsed.Address)

			if i == 0 {
				require.Equal(t, "Sign in to Example App", *parsed.Statement)
			} else {
				require.Nil(t, parsed.Statement)
			}

			println(parsed.IssuedAt.String())
			require.Equal(t, "2025-11-03 10:55:43.367 +0000 UTC", parsed.IssuedAt.String())
			require.Equal(t, "https://example.com", parsed.URI.String())
			require.Equal(t, "kaspa_mainnet", parsed.NetworkID)
			require.Equal(t, "12345678", parsed.Nonce)
			// require.Equal(t, "abcdef", *parsed.RequestID)

			ok, verifyErr := parsed.VerifySignature(example.signature)
			require.NoError(t, verifyErr)
			require.True(t, ok)
		})
	}
}
