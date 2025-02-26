package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	assert.Empty(t, cf.Token, "tokens must be empty")
	assert.Empty(t, cf.DefaultShell, "default shell must be empty")
	assert.Empty(t, cf.DefaultWorkDir, "default work dir must be empty")
	assert.NotEmpty(t, cf.BindAddress, "BindAddress should not be empty")
}

func TestTokens(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	cf.Token = append(cf.Token, Token{Token: "secret"})
	cf.Token = append(cf.Token, Token{Token: "secret2"})
	assert.False(t, cf.CheckToken(""), "empty token must be rejected")
	assert.False(t, cf.CheckToken("wrong"), "empty token must be rejected")
	assert.True(t, cf.CheckToken("secret"), "valid token must be accepted")
	assert.True(t, cf.CheckToken("secret2"), "valid token must be accepted")
}
