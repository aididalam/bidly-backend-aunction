package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port, DatabaseURL, JWTSecret, AWSRegion, S3Bucket, S3PublicBaseURL string
	S3Endpoint                                                         string
	S3UsePathStyle                                                     bool
	ExpireInterval                                                     time.Duration
}

func Load() (Config, error) { return load(os.Getenv) }

func load(getenv func(string) string) (Config, error) {
	c := Config{
		Port: strings.TrimSpace(getenv("AUCTION_SERVICE_PORT")), DatabaseURL: strings.TrimSpace(getenv("AUCTION_DATABASE_URL")),
		JWTSecret: getenv("JWT_SECRET"), AWSRegion: strings.TrimSpace(getenv("AWS_REGION")),
		S3Bucket: strings.TrimSpace(getenv("AWS_S3_BUCKET")), S3PublicBaseURL: strings.TrimRight(strings.TrimSpace(getenv("S3_PUBLIC_BASE_URL")), "/"),
		S3Endpoint:     strings.TrimRight(strings.TrimSpace(getenv("S3_ENDPOINT")), "/"),
		ExpireInterval: 30 * time.Second,
	}
	var missing []string
	for key, value := range map[string]string{"AUCTION_SERVICE_PORT": c.Port, "AUCTION_DATABASE_URL": c.DatabaseURL, "JWT_SECRET": c.JWTSecret, "AWS_REGION": c.AWSRegion, "AWS_S3_BUCKET": c.S3Bucket, "S3_PUBLIC_BASE_URL": c.S3PublicBaseURL} {
		if value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, errors.New("AUCTION_SERVICE_PORT must be a number from 1 to 65535")
	}
	if len(c.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	c.S3UsePathStyle, err = strconv.ParseBool(defaultValue(getenv("S3_USE_PATH_STYLE"), "false"))
	if err != nil {
		return Config{}, errors.New("S3_USE_PATH_STYLE must be true or false")
	}
	if value := strings.TrimSpace(getenv("EXPIRE_INTERVAL")); value != "" {
		c.ExpireInterval, err = time.ParseDuration(value)
		if err != nil || c.ExpireInterval <= 0 {
			return Config{}, errors.New("EXPIRE_INTERVAL must be a positive Go duration")
		}
	}
	return c, nil
}

func (c Config) Address() string { return ":" + c.Port }

func defaultValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
