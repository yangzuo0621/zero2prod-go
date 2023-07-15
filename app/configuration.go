package app

import (
	"fmt"

	"github.com/spf13/viper"
)

type Settings struct {
	Database        DatabaseSettings `mapstructure:"database"`
	ApplicationPort uint16           `mapstructure:"application_port"`
}

type DatabaseSettings struct {
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Port         uint16 `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	DatabaseName string `mapstructure:"database_name"`
}

func (settings *DatabaseSettings) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		settings.Username,
		settings.Password,
		settings.Host,
		settings.Port,
		settings.DatabaseName,
	)
}

func (settings *DatabaseSettings) ConnectionStringWithoutDB() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d?sslmode=disable",
		settings.Username,
		settings.Password,
		settings.Host,
		settings.Port,
	)
}

func GetConfiguration(path string) (*Settings, error) {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	settings := Settings{}
	err = viper.Unmarshal(&settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}
