package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var appConfig *Config

func GetAppConfig() *Config {
	return appConfig
}

func InitConfig() {
	env, ok := os.LookupEnv("ENVIRONMENT")
	if !ok {
		env = "dev"
		logrus.WithField("env", env).Infoln("运行环境 环境变量不存在，设置默认值")
	}
	configFilePath := getConfigFilePath()
	initial(fmt.Sprintf("%s/conf.%s.yml", configFilePath, env))
}

func initial(configFile string) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		logrus.WithField("config file path", configFile).Fatalf("读取配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(file, &appConfig)
	if err != nil {
		logrus.WithField("config file path", configFile).Fatalf("解析配置文件异常: %v", err)
	}
	logrus.WithField("config file path", configFile).Infoln("config initialized successfully")
}

func getConfigFilePath() string {
	configFilePath, ok := os.LookupEnv("CONFIG_FILE")
	if !ok {
		pwd, err := os.Getwd()
		configFilePath = fmt.Sprintf("%s/resources/config", pwd)
		if err != nil {
			logrus.Errorf("获取当前目录失败 %v", err)
			configFilePath = "./resources/config"
		}
		logrus.WithField("ConfigFilePath", configFilePath).Infoln("配置文件路径环境变量不存在，设置默认值")
	}
	return configFilePath
}
