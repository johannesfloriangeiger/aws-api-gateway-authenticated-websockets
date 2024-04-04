package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"math/big"
	"net/http"
)

type JsonWebKeys struct {
	Keys []JsonWebKey `json:"keys"`
}

type JsonWebKey struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

func getJsonWebKeys(jsonWebKeysUrl string) (*JsonWebKeys, error) {
	resp, err := http.Get(jsonWebKeysUrl)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	jsonWebKeys := new(JsonWebKeys)
	err = json.Unmarshal(body, jsonWebKeys)
	if err != nil {
		return nil, err
	}

	return jsonWebKeys, nil
}

func getPublicKeyForToken(jsonWebKeys *JsonWebKeys, token *jwt.Token) (*rsa.PublicKey, error) {
	jsonWebKey, err := getJsonWebKeyForToken(jsonWebKeys, token)
	if err != nil {
		return nil, err
	}

	return createPublicKey(*jsonWebKey)
}

const KeyId = "kid"

func getJsonWebKeyForToken(jsonWebKeys *JsonWebKeys, token *jwt.Token) (*JsonWebKey, error) {
	keyId := token.Header[KeyId]

	for _, key := range jsonWebKeys.Keys {
		if key.Kid == keyId {
			return &key, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("key %s not found", keyId))
}

func createPublicKey(key JsonWebKey) (*rsa.PublicKey, error) {
	modulus, err := decodeModulus(key)
	if err != nil {
		return nil, err
	}

	exponent, err := decodeExponent(key)
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: modulus,
		E: exponent,
	}, nil
}

func decodeModulus(key JsonWebKey) (*big.Int, error) {
	decodedModulus, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	modulus := big.NewInt(0)
	modulus.SetBytes(decodedModulus)

	return modulus, nil
}

const ExponentMinimumLength = 4

func decodeExponent(key JsonWebKey) (int, error) {
	decodedExponent, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return 0, err
	}

	if len(decodedExponent) < ExponentMinimumLength {
		sanitisedExponent := make([]byte, ExponentMinimumLength)
		copy(sanitisedExponent[ExponentMinimumLength-len(decodedExponent):], decodedExponent)
		decodedExponent = sanitisedExponent
	}

	return int(binary.BigEndian.Uint32(decodedExponent[:])), nil
}
