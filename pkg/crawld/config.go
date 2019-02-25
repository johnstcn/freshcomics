package crawld

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DSN                  string `default:"host=localhost user=freshcomics password=freshcomics_password dbname=freshcomicsdb sslmode=disable"`
	UserAgent            string `default:"freshcomics/crawld"`
	FetchTimeoutSecs     int    `default:"3"`
	CheckIntervalSecs    int    `default:"3600"`
	WorkPollIntervalSecs int    `default:"10"`
	ScheduleIntervalSecs int    `default:"60"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("crawld", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
