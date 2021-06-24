package config

type Config struct {
	Host string `default:"0.0.0.0"`
	Port int    `default:"3000"`
	DSN  string `default:"host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable"`
}
