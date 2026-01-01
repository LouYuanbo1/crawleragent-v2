package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitConfig() (*Config, error) {
	// 设置默认值
	viper.SetDefault("elasticsearch.host", "http://localhost")
	viper.SetDefault("elasticsearch.port", 9200)

	// 设置默认值
	viper.SetDefault("embedding.host", "http://localhost")
	viper.SetDefault("embedding.port", 11434)

	// 设置默认值
	viper.SetDefault("llm.host", "http://localhost")
	viper.SetDefault("llm.port", 11434)

	// 配置文件设置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	path := filepath.Join("..", "..", "config")
	//Viper查询的路径是相对于当前工作目录（current working directory） 的，而不是相对于可执行文件的位置或源代码文件的位置。
	// 1.尝试加载main.go所在目录的上两级目录下config.yaml
	viper.AddConfigPath(path)
	// 2.尝试加载当前工作目录下的config.yaml
	viper.AddConfigPath(".")

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// 设置绝对路径
	// 3.尝试加载main.go可执行文件所在目录的上两级目录下的config.yaml
	viper.AddConfigPath(filepath.Join(exeDir, "..", "..", "config"))
	// 4.尝试加载可执行文件所在目录下的config.yaml
	viper.AddConfigPath(exeDir)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found, using defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 绑定环境变量
	viper.AutomaticEnv()

	// 监控配置变化
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	viper.WatchConfig()

	cfg, err := parseConfig()
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	return cfg, nil
}

func parseConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}
	return &cfg, nil
}
