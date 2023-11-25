// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package keys_test

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/hdecarne-github/go-certstore/keys"
	"github.com/stretchr/testify/require"
)

func TestEqual(t *testing.T) {
	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	require.NoError(t, err)
	ed25519PublicKey, ed25519PrivateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// ecdsa
	require.True(t, keys.PrivatesEqual(ecdsaPrivateKey, ecdsaPrivateKey))
	require.False(t, keys.PrivatesEqual(ecdsaPrivateKey, ed25519PrivateKey))
	require.False(t, keys.PrivatesEqual(ecdsaPrivateKey, rsaPrivateKey))

	require.True(t, keys.PublicsEqual(&ecdsaPrivateKey.PublicKey, &ecdsaPrivateKey.PublicKey))
	require.False(t, keys.PublicsEqual(&ecdsaPrivateKey.PublicKey, ed25519PublicKey))
	require.False(t, keys.PublicsEqual(&ecdsaPrivateKey.PublicKey, &rsaPrivateKey.PublicKey))

	// ed25519
	require.False(t, keys.PrivatesEqual(ed25519PrivateKey, ecdsaPrivateKey))
	require.True(t, keys.PrivatesEqual(ed25519PrivateKey, ed25519PrivateKey))
	require.False(t, keys.PrivatesEqual(ed25519PrivateKey, rsaPrivateKey))

	require.False(t, keys.PublicsEqual(ed25519PublicKey, &ecdsaPrivateKey.PublicKey))
	require.True(t, keys.PublicsEqual(ed25519PublicKey, ed25519PublicKey))
	require.False(t, keys.PublicsEqual(ed25519PublicKey, &rsaPrivateKey.PublicKey))

	// rsa
	require.False(t, keys.PrivatesEqual(rsaPrivateKey, ecdsaPrivateKey))
	require.False(t, keys.PrivatesEqual(rsaPrivateKey, ed25519PrivateKey))
	require.True(t, keys.PrivatesEqual(rsaPrivateKey, rsaPrivateKey))

	require.False(t, keys.PublicsEqual(&rsaPrivateKey.PublicKey, &ecdsaPrivateKey.PublicKey))
	require.False(t, keys.PublicsEqual(&rsaPrivateKey.PublicKey, ed25519PublicKey))
	require.True(t, keys.PublicsEqual(&rsaPrivateKey.PublicKey, &rsaPrivateKey.PublicKey))
}

func TestProviders(t *testing.T) {
	for _, providerName := range keys.Providers() {
		providerKPFs := keys.ProviderKeyPairFactories(providerName)
		require.NotNil(t, providerKPFs)
		require.NotEqual(t, 0, len(providerKPFs))
		for _, providerKPF := range providerKPFs {
			kpf := keys.ProviderKeyPairFactory(providerKPF.Name())
			require.NotNil(t, kpf)
			require.Equal(t, providerKPF.Name(), kpf.Name())
		}
	}
}
