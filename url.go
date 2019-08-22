package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// MakeURL creates a redirector URL
func MakeURL(c Config, v map[string]interface{}) (string, error) {
	_, ok := v["url"]
	if !ok {
		return "", errors.New("A 'url' parameter must be provided")
	}

	// Remove and save the visible slug
	prefix := ""
	slug, hasSlug := v["slug"]
	if hasSlug && slug != "" {
		ss, ok := slug.(string)
		if ok {
			prefix = strings.Trim(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(ss, "-"), "-") + "/"
		}
	}
	delete(v, "slug")

	// Convert our payload to json
	plaintext, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	key, ok := c.Secrets[c.Version]
	if !ok {
		return "", errors.New("Malformed configuration")
	}

	// Encrypt our plaintext with the 128 bit AES-GCM
	// This hides the data and confirms it hasn't been altered.
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// base64 encode it using a URL-friendly encoding
	cookie := base64.RawURLEncoding.EncodeToString(ciphertext)

	// Using fmt.Sprintf rather than net/url as we're just
	// sample code. We know the dynamicaly created bits are
	// URL safe, so unless the config is broken we're OK

	return fmt.Sprintf("%s%s?x=%d.%s", c.BaseURL, prefix, c.Version, cookie), nil
}

// DecryptURL decrypts and validates a redirector URL
func DecryptURL(c Config, u *url.URL) (map[string]interface{}, error) {

	// Get the ?x= value from the URL
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	x := values.Get("x")
	if x == "" {
		return nil, errors.New("No x= parameter provided")
	}

	// Split it into version and ciphertext
	parts := strings.SplitN(x, ".", 2)
	if len(parts) != 2 {
		return nil, errors.New("Malformed x parameter")
	}
	version, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	key, ok := c.Secrets[int(version)]
	if !ok {
		return nil, fmt.Errorf("Unreocognized version: %d", version)
	}

	// Decode our ciphertext
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("Malformed ciphertext")
	}

	plaintext, err := gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)

	if err != nil {
		return nil, err
	}

	v := map[string]interface{}{}
	err = json.Unmarshal(plaintext, &v)
	return v, err
}
