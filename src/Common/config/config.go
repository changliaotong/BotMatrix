package config

// AppConfig represents the backend configuration
type AppConfig struct {
	WSPort               string `json:"ws_port"`
	WebUIPort            string `json:"webui_port"`
	RedisAddr            string `json:"redis_addr"`
	RedisPwd             string `json:"redis_pwd"`
	JWTSecret            string `json:"jwt_secret"`
	DefaultAdminPassword string `json:"default_admin_password"`
	StatsFile            string `json:"stats_file"`

	// Database Configuration
	PGHost     string `json:"pg_host"`
	PGPort     int    `json:"pg_port"`
	PGUser     string `json:"pg_user"`
	PGPassword string `json:"pg_password"`
	PGDBName   string `json:"pg_dbname"`
	PGSSLMode  string `json:"pg_sslmode"`

	// Feature Flags
	EnableSkill bool   `json:"enable_skill"`
	LogLevel    string `json:"log_level"`
	AutoReply   bool   `json:"auto_reply"`

	// Azure Translator Config
	AzureTranslateKey      string `json:"azure_translate_key"`
	AzureTranslateEndpoint string `json:"azure_translate_endpoint"`
	AzureTranslateRegion   string `json:"azure_translate_region"`
}

// ConnectionConfig represents a connection configuration
type ConnectionConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "v11", "v12", etc.
	Address  string `json:"address"`
	Token    string `json:"token"`
	Enabled  bool   `json:"enabled"`
	Platform string `json:"platform"`
}
