package services

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	UseMock    bool

	// PostgreSQL / Neon
	DatabaseURL string

	DBType     string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret          string
	JWTExpirationMin   int
	JWTRenewMin        int
	CookieName         string
	CookieSecure       bool
	CORSOrigin         string
	JWTStoreEnabled    bool
	JWTLoadUserOnRenew bool

	PublicDir            string
	XAccelRedirectPrefix string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		UseMock:            getEnvBool("USE_MOCK", true),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DBType:             getEnv("DB_TYPE", "postgres"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnvInt("DB_PORT", 5432),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "reservslot"),
		DBSSLMode:          getEnv("DB_SSLMODE", "require"),
		JWTSecret:          getEnv("JWT_SECRET", "cambia-este-secreto-en-produccion"),
		JWTExpirationMin:   getEnvInt("JWT_EXPIRATION_MINUTES", 60),
		JWTRenewMin:        getEnvInt("JWT_RENEW_MINUTES", 20),
		CookieName:         getEnv("COOKIE_NAME", "reservslot_token"),
		CookieSecure:       getEnvBool("COOKIE_SECURE", false),
		CORSOrigin:         getEnv("CORS_ORIGIN", "http://localhost:4200"),
		JWTStoreEnabled:    getEnvBool("JWT_STORE_ENABLED", true),
		JWTLoadUserOnRenew: getEnvBool("JWT_LOAD_USER_ON_RENEW", true),
		PublicDir:          getEnv("PUBLIC_DIR", "./public"),
		XAccelRedirectPrefix: getEnv("X_ACCEL_REDIRECT_PREFIX", ""),
	}

	return cfg, nil
}

// PublicPath es la carpeta pública servida por nginx (subruta "public" fija).
func (c *Config) PublicPath() string {
	return filepath.Join(c.PublicDir, "public")
}

// ImagesDir es el directorio donde viven las imágenes de ambientes.
func (c *Config) ImagesDir() string {
	return filepath.Join(c.PublicPath(), "imagenes")
}

// DBDSN devuelve la cadena de conexión a PostgreSQL.
// Si DATABASE_URL está definido (Neon) se usa tal cual; en caso contrario se
// compone a partir de las variables sueltas.
func (c *Config) DBDSN() string {
	if strings.TrimSpace(c.DatabaseURL) != "" {
		return c.DatabaseURL
	}
	switch c.DBType {
	case "postgres":
		return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + strconv.Itoa(c.DBPort) + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
	default:
		return ""
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
