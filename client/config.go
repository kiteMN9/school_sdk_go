package client

import (
	"school_sdk/config"
	"time"
)

type Config struct {
	baseURL     string
	timeout     time.Duration
	ExistVerify bool
	//Verify      string
	userAgent string
}

func NewConfig(baseURL string, existVerify bool, timeout time.Duration, userAgent string) *Config {
	if !config.CheckUALegal(userAgent) {
		userAgent = config.EdgeUA
	}
	return &Config{
		baseURL:     baseURL,
		timeout:     timeout,
		ExistVerify: existVerify,
		userAgent:   userAgent,
	}
}
