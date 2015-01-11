package f1le

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	RootPath string `json:"-"`
	Hash     string `json:"hash"`
	Files    []File `json:"files"`
	mutex    sync.RWMutex `json:"-"`
}

func LoadConfig(rootPath string) (*Config, error) {
	cfgPath := filepath.Join(rootPath, "config.json")
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		// Create a default configuration with password "password".
		cfg := &Config{rootPath,
			"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
			[]File{}, sync.RWMutex{}}
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

func (c *Config) FilesCopy() []File {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make([]File, len(c.Files))
	copy(res, c.Files)
	return res
}

func (c *Config) Save() error {
	if data, err := json.Marshal(c); err != nil {
		return err
	} else {
		configPath := filepath.Join(c.RootPath, "config.json")
		return ioutil.WriteFile(configPath, data, os.FileMode(0700))
	}
}

func (c *Config) Upload(input io.Reader, remoteName string) error {
	fileId := RandomString()
	localPath := filepath.Join(c.RootPath, fileId)
	output, err := os.Create(localPath)
	if err != nil {
		return err
	}
	
	size, err := io.Copy(output, input)
	output.Close()
	if err != nil {
		os.Remove(localPath)
		return err
	}
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	file := File{remoteName, fileId, time.Now().UTC().UnixNano(), size}
	c.Files = append([]File{file}, c.Files...)
	return c.Save()
}

type File struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	Uploaded int64  `json:"uploaded"`
	Size     int64  `json:"size"`
}

func RandomString() string {
	randomNumber := strconv.Itoa(rand.Int()) + strconv.Itoa(rand.Int())
	hash := sha256.Sum256([]byte(randomNumber))
	return hex.EncodeToString(hash[:])
}
