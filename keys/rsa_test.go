// Copyright (C) 2023 Holger de Carne and contributors
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package keys_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hdecarne-github/go-certstore/keys"
	"github.com/stretchr/testify/require"
)

func TestRSAKeyPair(t *testing.T) {
	kpfs := keys.ProviderKeyPairFactories("RSA")
	for _, kpf := range kpfs {
		fmt.Printf("Generating %s", kpf.Name())
		start := time.Now()
		keypair, err := kpf.New()
		elapsed := time.Since(start)
		fmt.Printf(" (took: %s)\n", elapsed)
		require.NoError(t, err)
		require.NotNil(t, keypair)
	}
}