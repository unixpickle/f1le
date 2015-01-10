package f1le

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	RootPath string `json:"-"`
	Hash     string
}

func LoadConfig(rootPath string) (*Config, error) {
	cfgPath := filepath.Join(rootPath, "config.json")
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		// Create a default configuration with password "password".
		cfg := &Config{rootPath,
			"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"}
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	res := &Config{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Config) CheckPassword(password string) bool {
	hash := sha256.Sum256([]byte(password))
	return c.Hash == strings.ToLower(hex.EncodeToString(hash[:]))
}

func (c *Config) Save() error {
	if data, err := json.Marshal(c); err != nil {
		return err
	} else {
		configPath := filepath.Join(c.RootPath, "config.json")
		return ioutil.WriteFile(configPath, data, os.FileMode(0700))
	}
}
