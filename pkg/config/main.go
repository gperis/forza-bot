package config

import (
	"fmt"
	"github.com/spf13/viper"
)

func Load(configName string, rawVal interface{}) {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	err = viper.Unmarshal(rawVal)
	if err != nil {
		fmt.Printf("There is an error with the configuration file, %v", err)
	}
}
