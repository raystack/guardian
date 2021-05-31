package app

import (
	"fmt"
	"strings"

	"github.com/jeremywohl/flatten"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/spf13/viper"
)

// LoadConfig returns application configuration
func LoadConfig() *domain.Config {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file was not found. Env vars and defaults will be used")
		} else {
			panic(fmt.Errorf("fatal error reading config: %s", err))
		}
	}

	configKeys, err := getFlattenedStructKeys(domain.Config{})
	if err != nil {
		panic(err)
	}

	// Bind each conf fields to environment vars
	for key := range configKeys {
		err := viper.BindEnv(configKeys[key])
		if err != nil {
			panic(err)
		}
	}

	var config domain.Config
	defaults.SetDefaults(&config)

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal config to struct: %v\n", err))
	}
	return &config
}

func getFlattenedStructKeys(config domain.Config) ([]string, error) {
	var structMap map[string]interface{}
	err := mapstructure.Decode(config, &structMap)
	if err != nil {
		return nil, err
	}

	flat, err := flatten.Flatten(structMap, "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}

	return keys, nil
}
