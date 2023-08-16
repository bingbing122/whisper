package config

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
	"whisper/pkg/nacos"
)

var (
	GlobalConfig *Config
	LOLConfig    *LolConfig
)

type LolConfig struct {
	Lol  LolCfg  `yaml:"lol"`
	LolM LolmCfg `yaml:"lolm"`
	Cron CronCfg `yaml:"cron"`
}
type CronCfg struct {
	Time    string `yaml:"time"`
	ReBuild bool   `yaml:"rebuild"`
}
type LolCfg struct {
	Equipment string `yaml:"equipment"`
	Heroes    string `yaml:"heroes"`
	Hero      string `yaml:"hero"`
	Rune      string `yaml:"rune"`
	Skill     string `yaml:"skill"`
}
type LolmCfg struct {
	Equipment       string `yaml:"equipment"`
	Heroes          string `yaml:"heroes"`
	Hero            string `yaml:"hero"`
	Rune            string `yaml:"rune"`
	RuneType        string `yaml:"runeType"`
	Skill           string `yaml:"skill"`
	RecommendHeroes string `yaml:"recommendHeroes"`
}

type Config struct {
	App      AppCfg      `yaml:"app"`
	Database DatabaseCfg `yaml:"database"`
	Redis    RedisCfg    `yaml:"redis"`
	MQ       MQCfg       `yaml:"mq"`
	ES       ESCfg       `yaml:"es"`
	Log      LogCfg      `yaml:"log"`
}
type AppCfg struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
}
type DatabaseCfg struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	DB       string `yaml:"db"`
}
type RedisCfg struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
type MQCfg struct {
	Schema   string `yaml:"schema"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
type ESCfg struct {
	Host       string   `yaml:"host"`
	Port       string   `yaml:"port"`
	BuildIndex []string `yaml:"buildIndex"`
}
type LogCfg struct {
	LogLevel int    `yaml:"logLevel"`
	Path     string `yaml:"path"`
	SqlLog   string `yaml:"sqlLog"`
	EsLog    string `yaml:"esLog"`
}

func Init() {

	// 初始化Nacos
	nacos.Init()
	content, err := nacos.ConfigClient.GetConfig(vo.ConfigParam{
		DataId: nacos.NacosConfig.Nacos.DataID,
		Group:  nacos.NacosConfig.Nacos.Group,
	})

	err = yaml.Unmarshal([]byte(content), &GlobalConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to unmarshal nacos config: %s \n", err))
	}

	cfg, _ := json.Marshal(*GlobalConfig)
	fmt.Println(string(cfg))

	// lolconfig
	content, err = nacos.ConfigClient.GetConfig(vo.ConfigParam{
		DataId: nacos.NacosConfig.LOL.DataID,
		Group:  nacos.NacosConfig.LOL.Group,
	})

	err = yaml.Unmarshal([]byte(content), &LOLConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to unmarshal LOLConfig: %s \n", err))
	}

	cfg, _ = json.Marshal(*LOLConfig)
	fmt.Println(string(cfg))

	err = nacos.ConfigClient.ListenConfig(vo.ConfigParam{
		DataId: nacos.NacosConfig.LOL.DataID,
		Group:  nacos.NacosConfig.LOL.Group,
		OnChange: func(namespace, group, dataId, data string) {
			err = yaml.Unmarshal([]byte(content), &LOLConfig)
			fmt.Println(fmt.Sprintf("LOLConfig: %#v \n", LOLConfig))
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to unmarshal nacos config: %s \n", err))
			}
		},
	})
	if err != nil {
		panic(err)
	}
}
