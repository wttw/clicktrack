package main

import (
	"crypto/rand"
	"encoding/json"
	"os"
)

// Secret contains two secret keys, for encryption and signing

// Config contains the persistent state for the app
type Config struct {
	Version int
	Secrets map[int][]byte
	Listen  string
	BaseURL string
}

// Load a Config from file
func (c *Config) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(f)
	return dec.Decode(c)
}

// Save a Config to file
func (c Config) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// Rotate our AES keys
func (c *Config) Rotate() error {
	encKey := make([]byte, 16)
	_, err := rand.Read(encKey)
	if err != nil {
		return err
	}
	c.Version++
	c.Secrets[c.Version] = encKey
	return nil
}

// Init a config
func (c *Config) Init() error {
	*c = Config{
		Version: 0,
		Listen:  "127.0.0.1:3000",
		BaseURL: "http://127.0.0.1:3000/",
		Secrets: map[int][]byte{},
	}
	return c.Rotate()
}

// LoadOrInit will either load a config or create and save one
func (c *Config) LoadOrInit(filename string) error {
	err := c.Load(filename)
	if !os.IsNotExist(err) {
		return err
	}
	err = c.Init()
	if err != nil {
		return err
	}
	return c.Save(filename)
}
