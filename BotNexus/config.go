package main

import (
	"os"
)

// 配置常量
var (
	WS_PORT    = ":3001"
	WEBUI_PORT = ":5000"
	STATS_FILE = "stats.json"
	REDIS_ADDR = "192.168.0.126:6379"
	REDIS_PWD  = "redis_zsYik8"
)

func init() {
	if v := os.Getenv("WS_PORT"); v != "" {
		WS_PORT = v
	}
	if v := os.Getenv("WEBUI_PORT"); v != "" {
		WEBUI_PORT = v
	}
	if v := os.Getenv("STATS_FILE"); v != "" {
		STATS_FILE = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		REDIS_ADDR = v
	}
	if v := os.Getenv("REDIS_PWD"); v != "" {
		REDIS_PWD = v
	}
}
