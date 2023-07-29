package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/wyattjychen/hades/internal/pkg/model"
)

var defaultConfig *model.Config

func LoadConfig(nodeType, configFileName string) (*model.Config, error) {
	var cfg model.Config
	cfgPath := fmt.Sprintf("internal/pkg/config/%s_config/%s", nodeType, configFileName)

	fmt.Println("config file path is: ", cfgPath)
	vip := viper.New()
	vip.SetConfigFile(cfgPath)
	vip.SetConfigType("json")
	err := vip.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("get config file: %s failed! \n", err))
	}
	vip.WatchConfig()

	vip.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := vip.Unmarshal(&cfg); err != nil {
			fmt.Println(err)
		}
	})
	if err := vip.Unmarshal(&cfg); err != nil {
		fmt.Println(err)
	}
	fmt.Println("load config file success. ")
	defaultConfig = &cfg
	return &cfg, nil
}

func GetConfig() *model.Config {
	return defaultConfig
}
