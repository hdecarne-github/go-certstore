// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package certstore_test

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/hdecarne-github/go-certstore"
	"github.com/hdecarne-github/go-certstore/certs"
	"github.com/hdecarne-github/go-certstore/keys"
	"github.com/hdecarne-github/go-certstore/storage"
	"github.com/stretchr/testify/require"
)

const testVersionLimit storage.VersionLimit = 2
const testCacheTTL = time.Minute * 10
const testKeyAlg = keys.ECDSA256

func TestNewStore(t *testing.T) {
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	require.NotNil(t, registry)
	require.Equal(t, "Registry[memory://]", registry.Name())
}

func TestCreateCertificate(t *testing.T) {
	name := "TestCreateCertificate"
	user := name + "User"
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	factory := newTestRootCertificateFactory(name)
	createdName, err := registry.CreateCertificate(name, factory, user)
	require.NoError(t, err)
	require.Equal(t, name, createdName)
	entry, err := registry.Entry(createdName)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.True(t, entry.HasKey())
	entryKey := entry.Key(user)
	require.NotNil(t, entryKey)
	require.True(t, entry.HasCertificate())
	entryCertificate := entry.Certificate()
	require.NotNil(t, entryCertificate)
	require.True(t, entry.IsRoot())
	require.True(t, entry.CanIssue(x509.KeyUsageCertSign))
}

func TestCreateCertificateRequest(t *testing.T) {
	name := "TestCreateCertificateRequest"
	user := name + "User"
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	factory := newTestCertificateRequestFactory(name)
	createdName, err := registry.CreateCertificateRequest(name, factory, user)
	require.NoError(t, err)
	require.Equal(t, name, createdName)
	entry, err := registry.Entry(createdName)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.True(t, entry.HasKey())
	entryKey := entry.Key(user)
	require.NotNil(t, entryKey)
	require.True(t, entry.HasCertificateRequest())
	entryCertificate := entry.CertificateRequest()
	require.NotNil(t, entryCertificate)
}

func TestResetRevocationList(t *testing.T) {
	name := "TestResetRevocationList"
	user := name + "User"
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	certFactory := newTestRootCertificateFactory(name)
	createdName, err := registry.CreateCertificate(name, certFactory, user)
	require.NoError(t, err)
	entry, err := registry.Entry(createdName)
	require.NoError(t, err)
	require.False(t, entry.HasRevocationList())
	revocationListFactory := newTestRevocationListFactory()
	revocationList1, err := entry.ResetRevocationList(revocationListFactory, user)
	require.NoError(t, err)
	require.NotNil(t, revocationList1)
	entry, err = registry.Entry(createdName)
	require.NoError(t, err)
	require.True(t, entry.HasRevocationList())
	revocationList2 := entry.RevocationList()
	require.NotNil(t, revocationList2)
	require.Equal(t, revocationList1, revocationList2)
}

func TestAttributes(t *testing.T) {
	name := "TestAttributes"
	user := name + "User"
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	factory := newTestRootCertificateFactory(name)
	createdName, err := registry.CreateCertificate(name, factory, user)
	require.NoError(t, err)
	entry, err := registry.Entry(createdName)
	require.NoError(t, err)
	attributes := map[string]string{"Key": "Value"}
	err = entry.SetAttributes(attributes)
	require.NoError(t, err)
	require.Equal(t, attributes, entry.Attributes())
}

func TestMerge(t *testing.T) {
	path, err := os.MkdirTemp("", "TestMerge*")
	require.NoError(t, err)
	defer os.RemoveAll(path)
	backend, err := storage.NewFSStorage(path, testVersionLimit)
	require.NoError(t, err)
	registry, err := certstore.NewStore(backend, testCacheTTL)
	require.NoError(t, err)
	otherRegistry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	user := "TestMergeUser"
	populateTestStore(t, otherRegistry, user, 5)
	start := time.Now()
	err = registry.Merge(otherRegistry, user)
	require.NoError(t, err)
	elapsed := time.Since(start)
	fmt.Printf("%s merged once (took: %s)\n", registry.Name(), elapsed)
	checkStoreEntries(t, registry, 160, 5)
	start = time.Now()
	err = registry.Merge(otherRegistry, user)
	require.NoError(t, err)
	elapsed = time.Since(start)
	fmt.Printf("%s merged twice (took: %s)\n", registry.Name(), elapsed)
	checkStoreEntries(t, registry, 160, 5)
}

func TestEntries(t *testing.T) {
	path, err := os.MkdirTemp("", "TestEntries*")
	require.NoError(t, err)
	defer os.RemoveAll(path)
	backend, err := storage.NewFSStorage(path, testVersionLimit)
	require.NoError(t, err)
	registry, err := certstore.NewStore(backend, testCacheTTL)
	require.NoError(t, err)
	user := "TestEntriesUser"
	populateTestStore(t, registry, user, 10)
	checkStoreEntries(t, registry, 1120, 10)
}

func TestCertPools(t *testing.T) {
	registry, err := certstore.NewStore(storage.NewMemoryStorage(testVersionLimit), 0)
	require.NoError(t, err)
	user := "TestCertPoolsUser"
	populateTestStore(t, registry, user, 5)
	roots, intermediates, err := registry.CertPools()
	require.NoError(t, err)
	require.NotNil(t, roots)
	require.NotNil(t, intermediates)
	entries, err := registry.Entries()
	require.NoError(t, err)
	for {
		entry, err := entries.Next()
		require.NoError(t, err)
		if entry == nil {
			break
		}
		if entry.HasCertificate() {
			options := &x509.VerifyOptions{
				Roots:         roots,
				Intermediates: intermediates,
			}
			chains, err := entry.Certificate().Verify(*options)
			require.NoError(t, err)
			require.Equal(t, 1, len(chains))
			if entry.IsRoot() {
				require.Equal(t, 1, len(chains[0]))
			} else if entry.IsCA() {
				require.Equal(t, 2, len(chains[0]))
			} else {
				require.Equal(t, 3, len(chains[0]))
			}
		}
	}
}

func checkStoreEntries(t *testing.T, registry *certstore.Registry, total int, roots int) {
	entries, err := registry.Entries()
	require.NoError(t, err)
	totalCount := 0
	rootCount := 0
	start := time.Now()
	for {
		nextEntry, err := entries.Next()
		require.NoError(t, err)
		if nextEntry == nil {
			break
		}
		totalCount++
		if nextEntry.IsRoot() {
			rootCount++
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("%s entries listed (took: %s)\n", registry.Name(), elapsed)
	require.Equal(t, total, totalCount)
	require.Equal(t, roots, rootCount)
}

func populateTestStore(t *testing.T, registry *certstore.Registry, user string, count int) {
	start := time.Now()
	createTestRootEntries(t, registry, user, count)
	createTestRequestEntries(t, registry, user, count)
	elapsed := time.Since(start)
	fmt.Printf("%s populated (took: %s)\n", registry.Name(), elapsed)
}

func createTestRootEntries(t *testing.T, registry *certstore.Registry, user string, count int) {
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("root%d", i+1)
		factory := newTestRootCertificateFactory(name)
		createdName, err := registry.CreateCertificate(name, factory, user)
		require.NoError(t, err)
		require.Equal(t, name, createdName)
		entry, err := registry.Entry(createdName)
		require.NoError(t, err)
		_, err = entry.ResetRevocationList(newTestRevocationListFactory(), user)
		require.NoError(t, err)
		createTestIntermediateEntries(t, registry, createdName, user, count)
	}
}

func createTestIntermediateEntries(t *testing.T, registry *certstore.Registry, issuerName string, user string, count int) {
	issuerEntry, err := registry.Entry(issuerName)
	require.NoError(t, err)
	issuerCert := issuerEntry.Certificate()
	issuerKey := issuerEntry.Key(user)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s:intermediate%d", issuerName, i+1)
		factory := newTestIntermediateCertificateFactory(name, issuerCert, issuerKey)
		createdName, err := registry.CreateCertificate(name, factory, user)
		require.NoError(t, err)
		require.Equal(t, name, createdName)
		createTestLeafEntries(t, registry, createdName, user, count)
	}
}

func createTestLeafEntries(t *testing.T, registry *certstore.Registry, issuerName string, user string, count int) {
	issuerEntry, err := registry.Entry(issuerName)
	require.NoError(t, err)
	issuerCert := issuerEntry.Certificate()
	issuerKey := issuerEntry.Key(user)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s:leaf%d", issuerName, i+1)
		factory := newTestLeafCertificateFactory(name, issuerCert, issuerKey)
		createdName, err := registry.CreateCertificate(name, factory, user)
		require.NoError(t, err)
		require.Equal(t, name, createdName)
	}
}

func createTestRequestEntries(t *testing.T, registry *certstore.Registry, user string, count int) {
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("request%d", i+1)
		factory := newTestCertificateRequestFactory(name)
		createdName, err := registry.CreateCertificateRequest(name, factory, user)
		require.NoError(t, err)
		require.Equal(t, name, createdName)
		_, err = registry.Entry(createdName)
		require.NoError(t, err)
	}
}

func newTestRootCertificateFactory(cn string) certs.CertificateFactory {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1),
	}
	return certs.NewLocalCertificateFactory(template, testKeyAlg.NewKeyPairFactory(), nil, nil)
}

func newTestIntermediateCertificateFactory(cn string, parent *x509.Certificate, signer crypto.PrivateKey) certs.CertificateFactory {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		KeyUsage:              x509.KeyUsageCertSign,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1),
	}
	return certs.NewLocalCertificateFactory(template, testKeyAlg.NewKeyPairFactory(), parent, signer)
}

func newTestLeafCertificateFactory(cn string, parent *x509.Certificate, signer crypto.PrivateKey) certs.CertificateFactory {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		BasicConstraintsValid: true,
		IsCA:                  false,
		MaxPathLen:            -1,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1),
	}
	return certs.NewLocalCertificateFactory(template, testKeyAlg.NewKeyPairFactory(), parent, signer)
}

func newTestCertificateRequestFactory(cn string) certs.CertificateRequestFactory {
	template := &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: cn},
	}
	return certs.NewRemoteCertificateRequestFactory(template, testKeyAlg.NewKeyPairFactory())
}

func newTestRevocationListFactory() certs.RevocationListFactory {
	now := time.Now()
	template := &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: now,
		NextUpdate: now.AddDate(0, 1, 0),
	}
	return certs.NewLocalRevocationListFactory(template)
}
