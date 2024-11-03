package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

var (
	config = ConfigStruct{}
	once   sync.Once

	configPath    = "conf.toml"
	defaultConfig = ConfigStruct{
		DataBase: DataBaseConfigStruct{
			Username:     "root",
			Host:         "127.0.0.1",
			Port:         3306,
			DatabaseName: "ACMBot",
			DriverName:   "MariaDB",
			Password:     "",
		},
		RWS: ReverseWebsocketConfigStruct{
			Host:        "0.0.0.0",
			Port:        5140,
			ChannelSize: 1000,
		},
	}
)

type ReverseWebsocketConfigStruct struct {
	Host        string `toml:"host"`
	Port        int32  `toml:"port"`
	Token       string `toml:"token"`
	ChannelSize int32  `toml:"channel_size"`
}

type DataBaseConfigStruct struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	DatabaseName string `toml:"database_name"`
	DriverName   string `toml:"driver_name"`
	AutoCreateDB bool   `toml:"auto_create_db"`
}

type CodeforcesConfigStruct struct {
	Key    string `toml:"key"`
	Secret string `toml:"secret"`
}

type ConfigStruct struct {
	Codeforces CodeforcesConfigStruct       `toml:"Codeforces"`
	DataBase   DataBaseConfigStruct         `toml:"DataBase"`
	RWS        ReverseWebsocketConfigStruct `toml:"ReverseWebsocket"`
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

func GetConfig() *ConfigStruct {
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

func Init() {
	GetConfig()
}
