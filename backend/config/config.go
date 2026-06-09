package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type RodConfig struct {
	UserDataDir          string `mapstructure:"user_data_dir"`
	UserMode             bool   `mapstructure:"user_mode"`
	Headless             bool   `mapstructure:"headless"`
	DisableBlinkFeatures string `mapstructure:"disable_blink_features"`
	Incognito            bool   `mapstructure:"incognito"`
	DisableDevShmUsage   bool   `mapstructure:"disable_dev_shm_usage"`
	NoSandbox            bool   `mapstructure:"no_sandbox"`
	DefaultPageWidth     int    `mapstructure:"default_page_width"`
	DefaultPageHeight    int    `mapstructure:"default_page_height"`
	UserAgent            string `mapstructure:"user_agent"`
	Leakless             bool   `mapstructure:"leakless"`
	Bin                  string `mapstructure:"bin"`
	//(禁用后台标签页定时器节流)
	DisableBackgroundNetworking bool `mapstructure:"disable_background_networking"`
	//(禁用后台网络) 设为false
	DisableBackgroundTimerThrottling bool `mapstructure:"disable_background_timer_throttling"`
	//(禁用后台窗口遮挡)
	DisableBackgroundingOccludedWindows bool `mapstructure:"disable_backgrounding_occluded_windows"`
	//(禁用渲染器后台)
	DisableRendererBackgrounding bool `mapstructure:"disable_renderer_backgrounding"`
	//(远程调试端口)
	BasicRemoteDebuggingPort int `mapstructure:"basic_remote_debugging_port"`
	//(开启CDP通信追踪)
	Trace bool `mapstructure:"trace"`
}

type DeepseekConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type EmbeddingConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Model string `mapstructure:"model"`
}

type ElasticsearchConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

type LLMConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Model string `mapstructure:"model"`
}

type PromptConfig struct {
	PromptDir string `mapstructure:"prompt_dir"`
}

type Config struct {
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`

	Rod RodConfig `mapstructure:"rod"`

	Embedding EmbeddingConfig `mapstructure:"embedding"`

	Deepseek DeepseekConfig `mapstructure:"deepseek"`

	LLM LLMConfig `mapstructure:"llm"`

	Prompt PromptConfig `mapstructure:"prompt"`
}

func InitConfig() (*Config, error) {
	// 配置文件设置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	path0 := filepath.Join(".")
	path1 := filepath.Join("config")
	path2 := filepath.Join("..", "config")
	path3 := filepath.Join("..", "..", "config")
	path4 := filepath.Join("..", "..", "..", "config")
	//Viper查询的路径是相对于当前工作目录（current working directory） 的，而不是相对于可执行文件的位置或源代码文件的位置。
	// 1.尝试加载main.go所在目录的上两级目录下config.yaml
	for _, path := range []string{path0, path1, path2, path3, path4} {
		viper.AddConfigPath(path)
	}
	// 2.尝试加载当前工作目录下的config.yaml
	viper.AddConfigPath(".")

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// 设置绝对路径
	// 3.尝试加载main.go可执行文件所在目录的上两级目录下的config.yaml
	pathExe0 := filepath.Join(exeDir)
	pathExe1 := filepath.Join(exeDir, "config")
	pathExe2 := filepath.Join(exeDir, "..", "config")
	pathExe3 := filepath.Join(exeDir, "..", "..", "config")
	// 4.尝试加载main.go可执行文件所在目录的上三级目录下的config.yaml
	pathExe4 := filepath.Join(exeDir, "..", "..", "..", "config")
	for _, path := range []string{pathExe0, pathExe1, pathExe2, pathExe3, pathExe4} {
		viper.AddConfigPath(path)
	}

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
