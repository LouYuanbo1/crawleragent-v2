package config

type Config struct {
	Elasticsearch struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Address  string `json:"address"`
	} `json:"elasticsearch"`

	Rod struct {
		UserDataDir          string `json:"user_data_dir"`
		UserMode             bool   `json:"user_mode"`
		Headless             bool   `json:"headless"`
		DisableBlinkFeatures string `json:"disable_blink_features"`
		Incognito            bool   `json:"incognito"`
		DisableDevShmUsage   bool   `json:"disable_dev_shm_usage"`
		NoSandbox            bool   `json:"no_sandbox"`
		DefaultPageWidth     int    `json:"default_page_width"`
		DefaultPageHeight    int    `json:"default_page_height"`
		UserAgent            string `json:"user_agent"`
		Leakless             bool   `json:"leakless"`
		Bin                  string `json:"bin"`
		//(禁用后台标签页定时器节流)
		DisableBackgroundNetworking bool `json:"disable_background_networking"`
		//(禁用后台网络) 设为false
		DisableBackgroundTimerThrottling bool `json:"disable-background-timer-throttling"`
		//(禁用后台窗口遮挡)
		DisableBackgroundingOccludedWindows bool `json:"disable-backgrounding-occluded-windows"`
		//(禁用渲染器后台)
		DisableRendererBackgrounding bool `json:"disable-renderer-backgrounding"`
		//(远程调试端口)
		BasicRemoteDebuggingPort int `json:"basic_remote_debugging_port"`
		//(开启CDP通信追踪)
		Trace bool `json:"trace"`
	} `json:"rod"`

	Embedding struct {
		Host  string `json:"host"`
		Port  int    `json:"port"`
		Model string `json:"model"`
	} `json:"embedding"`

	LLM struct {
		Host  string `json:"host"`
		Port  int    `json:"port"`
		Model string `json:"model"`
	} `json:"llm"`
}
