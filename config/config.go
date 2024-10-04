package db

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	DatabaseType string   `json:"database_type"`
	Postgres     DBConfig `json:"postgres"`
	MySQL        DBConfig `json:"mysql"`
}

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	Port     int    `json:"port"`
	SSLMode  string `json:"sslmode,omitempty"`
}

func LoadDBConfig(filepath string) (*Config, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Error reading database config file: ", err)
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Error parsing database config file: ", err)
		return nil, err
	}

	return &config, nil
}
