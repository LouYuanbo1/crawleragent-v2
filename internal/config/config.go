package config

type Config struct {
	Elasticsearch struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
	} `mapstructure:"elasticsearch"`

	Rod struct {
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
	} `mapstructure:"rod"`

	Embedding struct {
		Host  string `mapstructure:"host"`
		Port  int    `mapstructure:"port"`
		Model string `mapstructure:"model"`
	} `mapstructure:"embedding"`

	LLM struct {
		Host  string `mapstructure:"host"`
		Port  int    `mapstructure:"port"`
		Model string `mapstructure:"model"`
	} `mapstructure:"llm"`
}
