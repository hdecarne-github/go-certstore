// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package store_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"

	store "github.com/hdecarne-github/go-certstore"
	"github.com/hdecarne-github/go-certstore/certs"
	"github.com/hdecarne-github/go-certstore/keys"
	"github.com/hdecarne-github/go-certstore/storage"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	registry, err := store.NewStore(storage.NewMemoryStorage(2))
	require.NoError(t, err)
	require.NotNil(t, registry)
}

func TestCreateCertificate(t *testing.T) {
	factory := newCertificateFactory()
	registry, err := store.NewStore(storage.NewMemoryStorage(2))
	require.NoError(t, err)
	name := "TestCreateCertificate"
	user := name + "User"
	createdName, err := registry.CreateCertificate(name, factory, user)
	require.NoError(t, err)
	require.Equal(t, name, createdName)
	entry, err := registry.Entry(createdName)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.True(t, entry.HasKey())
	entryKey, err := entry.Key()
	require.NoError(t, err)
	require.NotNil(t, entryKey)
	require.True(t, entry.HasCertificate())
	entryCertificate, err := entry.Certificate()
	require.NoError(t, err)
	require.NotNil(t, entryCertificate)
}

func newCertificateFactory() certs.CertificateFactory {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63n(math.MaxInt64)),
		Subject: pkix.Name{
			Organization: []string{now.Local().String()},
		},
		NotBefore: now,
		NotAfter:  now.Add(time.Hour),
	}
	return certs.NewLocalCertificateFactory(template, keys.ProviderKeyPairFactory("ECDSA P-224"), nil, nil)
}
