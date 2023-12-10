// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package certs

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"math/big"

	"github.com/hdecarne-github/go-certstore/keys"
	"github.com/hdecarne-github/go-log"
	"github.com/rs/zerolog"
)

const localCertificateFactoryName = "Local"

type localCertificateFactory struct {
	template       *x509.Certificate
	keyPairFactory keys.KeyPairFactory
	parent         *x509.Certificate
	signer         crypto.PrivateKey
	logger         *zerolog.Logger
}

func (factory *localCertificateFactory) Name() string {
	return localCertificateFactoryName
}

func (factory *localCertificateFactory) New() (crypto.PrivateKey, *x509.Certificate, error) {
	keyPair, err := factory.keyPairFactory.New()
	if err != nil {
		return nil, nil, err
	}
	createTemplate := factory.template
	var certificateBytes []byte
	if factory.parent != nil {
		// parent signed
		factory.logger.Info().Msg("creating signed local X.509 certificate...")
		createTemplate.SerialNumber = nextSerialNumber()
		certificateBytes, err = x509.CreateCertificate(rand.Reader, createTemplate, createTemplate, keyPair.Public(), factory.signer)
	} else {
		// self-signed
		factory.logger.Info().Msg("creating self-signed local X.509 certificate...")
		createTemplate.SerialNumber = big.NewInt(1)
		certificateBytes, err = x509.CreateCertificate(rand.Reader, createTemplate, createTemplate, keyPair.Public(), keyPair.Private())
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate (cause: %w)", err)
	}
	certificate, err := x509.ParseCertificate(certificateBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parse certificate bytes (cause: %w)", err)
	}
	return keyPair.Private(), certificate, nil
}

// NewLocalCertificateFactory creates a new certificate factory for locally issued certificates.
func NewLocalCertificateFactory(template *x509.Certificate, keyPairFactory keys.KeyPairFactory, parent *x509.Certificate, signer crypto.PrivateKey) CertificateFactory {
	logger := log.RootLogger().With().Str("CertificateFactory", localCertificateFactoryName).Logger()
	return &localCertificateFactory{
		template:       template,
		keyPairFactory: keyPairFactory,
		parent:         parent,
		signer:         signer,
		logger:         &logger,
	}
}

type localRevocationListFactory struct {
	template *x509.RevocationList
	issuer   *x509.Certificate
	signer   crypto.PrivateKey
	logger   *zerolog.Logger
}

func (factory *localRevocationListFactory) Name() string {
	return localCertificateFactoryName
}

func (factory *localRevocationListFactory) New() (*x509.RevocationList, error) {
	factory.logger.Info().Msg("creating local X.509 revocation list...")
	revocationListBytes, err := x509.CreateRevocationList(rand.Reader, factory.template, factory.issuer, keys.KeyFromPrivate(factory.signer))
	if err != nil {
		return nil, fmt.Errorf("failed to create revocation list (cause: %w)", err)
	}
	revocationList, err := x509.ParseRevocationList(revocationListBytes)
	if err != nil {
		return nil, fmt.Errorf("failed parse revocation list bytes (cause: %w)", err)
	}
	return revocationList, nil
}

// NewLocalRevocationListFactory creates a new revocation list factory for locally issued certificates.
func NewLocalRevocationListFactory(template *x509.RevocationList, issuer *x509.Certificate, signer crypto.PrivateKey) RevocationListFactory {
	logger := log.RootLogger().With().Str("Factory", localCertificateFactoryName).Logger()
	return &localRevocationListFactory{
		template: template,
		issuer:   issuer,
		signer:   signer,
		logger:   &logger,
	}
}
