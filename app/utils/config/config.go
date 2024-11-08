package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	config = ConfigStruct{}
	once   sync.Once

	configPath    = "./conf.toml"
	configName    = "conf"
	defaultConfig = ConfigStruct{
		DataBase: DataBaseConfigStruct{
			Username:     "root",
			Host:         "127.0.0.1",
			Port:         3306,
			DatabaseName: "ACMBot",
			DriverName:   "MariaDB",
			Password:     "",
		},
		Bot: BotConfigStruct{
			NickName:      []string{"bot1", "bot2"},
			CommandPrefix: "#",
			SuperUsers:    []int64{123456789, 987654321},
			WS: []WebsocketConfigStruct{
				{Host: "localhost",
					Port: 3001},
			},
		},
	}
)

type WebsocketConfigStruct struct {
	IsForward   bool   `mapstructure:"is_forward" toml:"is_forward"`
	Host        string `mapstructure:"host" toml:"host"`
	Port        int32  `mapstructure:"port" toml:"port"`
	Token       string `mapstructure:"token" toml:"token"`
	ChannelSize int    `mapstructure:"channle_size" toml:"channel_size"`
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

type ConfigStruct struct {
	Codeforces CodeforcesConfigStruct `mapstructure:"codeforces" toml:"codeforces"`
	DataBase   DataBaseConfigStruct   `mapstructure:"database" toml:"database"`
	Bot        BotConfigStruct        `mapstructure:"bot" toml:"bot"`
}

func writeConfig(config *ConfigStruct) error {
	// create

	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			log.Errorf("failed to close config file: %v", err)
		}
	}(configFile)

	// write
	if err := toml.NewEncoder(configFile).Encode(config); err != nil {
		return fmt.Errorf("failed to write default config: %v", err)
	}
	return nil
}

func overwriteConfig(defaultConfig, actualConfig interface{}) error {
	return overwriteConfigRecursive(reflect.ValueOf(defaultConfig).Elem(), reflect.ValueOf(actualConfig).Elem())
}

func overwriteConfigRecursive(defaultVal, actualVal reflect.Value) error {
	if defaultVal.Kind() != reflect.Struct || actualVal.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < actualVal.NumField(); i++ {
		actualField := actualVal.Field(i)
		if actualField.CanInterface() {
			fieldName := actualVal.Type().Field(i).Tag.Get("toml")
			if fieldName == "" {
				continue
			}

			defaultField := defaultVal.FieldByName(actualVal.Type().Field(i).Name)
			if defaultField.IsValid() && defaultField.CanSet() {
				if actualField.Kind() == reflect.Struct && defaultField.Kind() == reflect.Struct {
					if err := overwriteConfigRecursive(defaultField, actualField); err != nil {
						return err
					}
				} else if !reflect.DeepEqual(actualField.Interface(), reflect.Zero(actualField.Type()).Interface()) {
					defaultField.Set(actualField)
				}
			}
		}
	}
	return nil
}

func GetConfig_() *ConfigStruct {
	once.Do(func() {
		configFile, err := os.Open(configPath)
		if os.IsNotExist(err) {
			log.Warnf("Config file not found, creating")
			err := writeConfig(&defaultConfig)
			if err != nil {
				log.Fatalf("Failed to create default config: %v", err)
			}
			log.Warnf("Default config %s created. Please complete it and restart me", configPath)
			os.Exit(0)
		} else if err != nil {
			log.Fatalf("failed to load config file %s: %v", configPath, err)
		}
		log.Debug("load file successfully, loading config")
		defer func(configFile *os.File) {
			err := configFile.Close()
			if err != nil {
				log.Errorf("failed to close config file: %v", err)
			}
		}(configFile)
		metaInFile, err := toml.NewDecoder(configFile).Decode(&config)
		if err != nil {
			log.Fatalf("failed to decode config file: %v", err)
		}
		var buf bytes.Buffer
		err = toml.NewEncoder(&buf).Encode(defaultConfig)
		if err != nil {
			log.Fatalf("(Unexpected branching, maybe there are something wrong in configStruct)failed to encode config: %v", err)
		}
		var mixConfig ConfigStruct
		metaInDefault, err := toml.NewDecoder(bytes.NewBufferString(buf.String())).Decode(&mixConfig)
		if err != nil {
			log.Fatalf("(Unexpected branching, maybe there are something wrong in model github.com/sirupsen/logrus)failed to decode config: %v", err)
		}

		keySet := make(map[string]bool)
		for _, v := range metaInDefault.Keys() {
			keySet[v.String()] = true
		}
		for _, v := range metaInFile.Keys() {
			delete(keySet, v.String())
		}
		if len(keySet) > 0 {
			log.Warnf("Some configuration is missing")
			log.Warnf("===================")
			for k := range keySet {
				log.Warnf(k)
			}
			log.Warnf("===================")

			if err = overwriteConfig(&mixConfig, &config); err != nil {
				log.Fatalf("Failed to populate the configuration from the file into the default configuration: %v", err)
			}
			log.Info("Overwriting from actual configuration to default configuration is done, no errors found")
			err = writeConfig(&mixConfig)
			if err != nil {
				log.Fatalf("Failed to write mixed configuration")
			}
			log.Info("Successfully written mixed configuration")
			log.Warnf("The default configuration may not be the expected configuration, please check the configuration")
		}
	})

	return &config
}

func configInit() {
	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.SetDefault("DataBase", defaultConfig.DataBase)
	viper.SetDefault("Bot", defaultConfig.Bot)
	viper.SetDefault("Codeforces", defaultConfig.Codeforces)
}

func configWrite() error {
	return viper.SafeWriteConfigAs(configPath)
}

func GetConfig() *ConfigStruct {
	once.Do(func() {
		configInit()
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Warnf("Config file not found, creating")
				if err := configWrite(); err != nil {
					log.Panicf("Config create failed: %v", err)
				}
				log.Warnf("Default config %s created. Please complete it and restart me", configPath)
				os.Exit(0)
			} else {
				log.Panicf("Reading config failed, when config file exist: %v", err)
			}
		}
		if err := viper.Unmarshal(&config); err != nil {
			log.Panicf("Unable to decode config into struct: %v", err)
		}
	})
	return &config
}
