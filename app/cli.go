package app

var (
	CLIConfigFileName      = ".guardian"
	CLIConfigFileExtension = "yaml"
)

type CLIConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
}

func LoadCLIConfig() (*CLIConfig, error) {
	var config CLIConfig
	if err := loadConfig(CLIConfigFileName, CLIConfigFileExtension, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
