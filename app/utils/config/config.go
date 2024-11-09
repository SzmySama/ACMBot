package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	config = AllConfigStruct{}
	once   sync.Once

	configPath     = "."
	configName     = "conf"
	configType     = "toml"
	fullConfigPath = fmt.Sprintf("%s/%s.%s", configPath, configName, configType)

	defaultConfig = AllConfigStruct{
		DataBase: DataBaseConfigStruct{
			Username:     "root",
			Host:         "127.0.0.1",
			Port:         3306,
			DatabaseName: "ACMBot",
			DriverName:   "MariaDB",
			Password:     "",
		},
		Bot: BotConfigStruct{
			NickName:      []string{"bot"},
			CommandPrefix: "#",
			SuperUsers:    []int64{},
			WS: []WebsocketConfigStruct{
				{
					Host: "localhost",
					Port: 3001,
				},
			},
		},
		Redis: RedisConfigStruct{
			Host:     "127.0.0.1",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
)

type WebsocketConfigStruct struct {
	IsForward   bool   `mapstructure:"is_forward" toml:"is_forward"`
	Host        string `mapstructure:"host" toml:"host"`
	Port        int32  `mapstructure:"port" toml:"port"`
	Token       string `mapstructure:"token" toml:"token"`
	ChannelSize int    `mapstructure:"channel_size" toml:"channel_size"`
}

type BotConfigStruct struct {
	NickName      []string                `mapstructure:"nick_name" toml:"nick_name"`
	CommandPrefix string                  `mapstructure:"command_prefix" toml:"command_prefix"`
	SuperUsers    []int64                 `mapstructure:"super_users" toml:"super_users"`
	WS            []WebsocketConfigStruct `mapstructure:"websocket" toml:"websocket"`
}

type CodeforcesConfigStruct struct {
	Key    string `mapstructure:"key" toml:"key"`
	Secret string `mapstructure:"secret" toml:"secret"`
}

type DataBaseConfigStruct struct {
	Host         string `mapstructure:"host" toml:"host"`
	Port         int    `mapstructure:"port" toml:"port"`
	Username     string `mapstructure:"username" toml:"username"`
	Password     string `mapstructure:"password" toml:"password"`
	DatabaseName string `mapstructure:"database_name" toml:"database_name"`
	DriverName   string `mapstructure:"driver_name" toml:"driver_name"`
	AutoCreateDB bool   `mapstructure:"auto_create_db" toml:"auto_create_db"`
}

type RedisConfigStruct struct {
	Host     string `mapstructure:"host" toml:"host"`
	Port     int    `mapstructure:"port" toml:"port"`
	Password string `mapstructure:"password" toml:"password"`
	DB       int    `mapstructure:"db" toml:"db"`
}

type AllConfigStruct struct {
	Bot        BotConfigStruct        `mapstructure:"bot" toml:"bot"`
	Codeforces CodeforcesConfigStruct `mapstructure:"codeforces" toml:"codeforces"`
	DataBase   DataBaseConfigStruct   `mapstructure:"database" toml:"database"`
	Redis      RedisConfigStruct      `mapstructure:"redis" toml:"redis"`
}

func configInit() {
	viper.SetConfigFile(fullConfigPath)
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	viper.SetDefault("DataBase", defaultConfig.DataBase)
	viper.SetDefault("Bot", defaultConfig.Bot)
	viper.SetDefault("Codeforces", defaultConfig.Codeforces)
	viper.SetDefault("Redis", defaultConfig.Redis)
}

func GetConfig() *AllConfigStruct {
	once.Do(func() {
		configInit()
		if err := viper.ReadInConfig(); err != nil {
			if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
				log.Panicf("Reading config failed, when config file exist: %v", err)
			}
			log.Warnf("Config file not found, creating")
			if err := viper.SafeWriteConfigAs(fullConfigPath); err != nil {
				log.Panicf("Config create failed: %v", err)
			}
			log.Warnf("Default config %s created. Please complete it and restart me", fullConfigPath)
			os.Exit(0)
		}
		if err := viper.Unmarshal(&config); err != nil {
			log.Panicf("Unable to decode config into struct: %v", err)
		}
	})
	return &config
}
