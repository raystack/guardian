package app

import (
	"fmt"
	"strings"

	"github.com/jeremywohl/flatten"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func loadConfig(fileName, fileExtension string, v interface{}) error {
	viper.SetConfigName(fileName)
	viper.SetConfigType(fileExtension)
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file was not found. Env vars and defaults will be used. If you are using guardian CLI, make sure you have run \"guardian config init\" once")
		} else {
			return fmt.Errorf("fatal error reading config: %s", err)
		}
	}

	configKeys, err := getFlattenedStructKeys(v)
	if err != nil {
		return err
	}

	// Bind each conf fields to environment vars
	for key := range configKeys {
		err := viper.BindEnv(configKeys[key])
		if err != nil {
			return err
		}
	}

	defaults.SetDefaults(v)

	if err := viper.Unmarshal(v); err != nil {
		return fmt.Errorf("unable to unmarshal config to struct: %v", err)
	}

	return nil
}

func getFlattenedStructKeys(v interface{}) ([]string, error) {
	var structMap map[string]interface{}
	err := mapstructure.Decode(v, &structMap)
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
