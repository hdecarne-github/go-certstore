// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package acme

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/hdecarne-github/go-certstore/keys"
)

type ProviderRegistration struct {
	Provider     string `json:"provider"`
	Email        string `json:"email"`
	EncodedKey   string `json:"key"`
	Registration *registration.Resource
}

func (providerRegistration *ProviderRegistration) GetEmail() string {
	return providerRegistration.Email
}

func (providerRegistration *ProviderRegistration) GetRegistration() *registration.Resource {
	return providerRegistration.Registration
}

func (providerRegistration *ProviderRegistration) GetPrivateKey() crypto.PrivateKey {
	if providerRegistration.EncodedKey == "" {
		return nil
	}
	keyBytes, err := base64.StdEncoding.DecodeString(providerRegistration.EncodedKey)
	if err != nil {
		return nil
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil
	}
	return key
}

func (providerRegistration *ProviderRegistration) matches(providerRegistration2 *ProviderRegistration) bool {
	return providerRegistration.Provider == providerRegistration2.Provider && providerRegistration.Email == providerRegistration2.Email
}

func (providerRegistration *ProviderRegistration) isActive(client *lego.Client) bool {
	if providerRegistration.Registration == nil {
		return false
	}
	_, err := client.Registration.QueryRegistration()
	return err == nil
}

func (providerRegistration *ProviderRegistration) register(client *lego.Client, keyFactory keys.KeyPairFactory) error {
	options := registration.RegisterOptions{TermsOfServiceAgreed: true}
	registrationResource, err := client.Registration.Register(options)
	if err != nil {
		return fmt.Errorf("failed to register at ACME provider '%s' (cause: %w)", providerRegistration.Provider, err)
	}
	providerRegistration.Registration = registrationResource
	return nil
}

func (providerRegistration *ProviderRegistration) updateProviderRegistrations(file *os.File) error {
	fileProviderRegistrations, err := unmarshalProviderRegistrations(file)
	if err != nil {
		return err
	}
	updateIndex := -1
	for i, fileProviderRegistration := range fileProviderRegistrations {
		if fileProviderRegistration.matches(providerRegistration) {
			updateIndex = i
			break
		}
	}
	if updateIndex >= 0 {
		fileProviderRegistrations[updateIndex] = *providerRegistration
	} else {
		fileProviderRegistrations = append(fileProviderRegistrations, *providerRegistration)
	}
	writeBytes, err := json.MarshalIndent(fileProviderRegistrations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registrations (cause: %w)", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek failed for file '%s' (cause: %w)", file.Name(), err)
	}
	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncate failed for file '%s' (cause: %w)", file.Name(), err)
	}
	_, err = file.Write(writeBytes)
	if err != nil {
		return fmt.Errorf("write failed for file '%s' (cause: %w)", file.Name(), err)
	}
	return nil
}

func prepareProviderRegistration(provider *ProviderConfig, file *os.File, keyPairFactory keys.KeyPairFactory) (*ProviderRegistration, error) {
	registrations, err := unmarshalProviderRegistrations(file)
	if err != nil {
		return nil, err
	}
	for _, registration := range registrations {
		if registration.Provider == provider.Name {
			return &registration, nil
		}
	}
	key, err := keyPairFactory.New()
	if err != nil {
		return nil, err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key.Private())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key (cause: %w)", err)
	}
	registration := &ProviderRegistration{
		Provider:   provider.Name,
		Email:      provider.RegistrationEmail,
		EncodedKey: base64.StdEncoding.EncodeToString(keyBytes),
	}
	return registration, nil
}

func unmarshalProviderRegistrations(file *os.File) ([]ProviderRegistration, error) {
	readBytes := make([]byte, 0, 4096)
	for {
		read, err := file.Read(readBytes)
		if read == 0 {
			break
		}
		if err != nil {
			return nil, err
		}
		readBytes = readBytes[:len(readBytes)+read]
		if len(readBytes) == cap(readBytes) {
			readBytes = append(readBytes, 0)[:len(readBytes)]
		}
	}
	registrations := make([]ProviderRegistration, 0)
	if len(readBytes) > 0 {
		err := json.Unmarshal(readBytes, &registrations)
		if err != nil {
			return nil, err
		}
	}
	return registrations, nil
}
