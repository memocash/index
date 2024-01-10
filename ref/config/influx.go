package config

type InfluxConfig struct {
	Url    string `mapstructure:"URL"`
	Org    string `mapstructure:"ORG"`
	Bucket string `mapstructure:"BUCKET"`
	Token  string `mapstructure:"TOKEN"`
	Source string `mapstructure:"SOURCE"`
}
