package crawler

import (
	"fmt"

	"github.com/go-rod/rod/lib/launcher"
)

// 定义选项函数类型
type LauncherOption func(*launcher.Launcher)

func CreateLauncher(userMode bool, options ...LauncherOption) *launcher.Launcher {
	var l *launcher.Launcher
	if userMode {
		l = launcher.NewUserMode()
	} else {
		l = launcher.New()
	}

	// 应用所有选项
	for _, option := range options {
		option(l)
	}

	return l
}

func WithUserDataDir(dir string) LauncherOption {
	return func(l *launcher.Launcher) {
		if dir != "" {
			l.Set("user-data-dir", dir)
		}
	}
}

func WithHeadless(headless bool) LauncherOption {
	return func(l *launcher.Launcher) {
		l.Headless(headless)
	}
}

func WithDisableBlinkFeatures(features string) LauncherOption {
	return func(l *launcher.Launcher) {
		if features != "" {
			l.Set("disable-blink-features", features)
		}
	}
}

func WithIncognito(incognito bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if incognito {
			l.Set("incognito")
		}
	}
}

func WithDisableDevShmUsage(disable bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if disable {
			l.Set("disable-dev-shm-usage")
		}
	}
}

func WithNoSandbox(noSandbox bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if noSandbox {
			l.Set("no-sandbox")
		}
	}
}

func WithLeakless(leakless bool) LauncherOption {
	return func(l *launcher.Launcher) {
		l.Leakless(leakless)
	}
}

// 各种选项函数
func WithBin(bin string) LauncherOption {
	return func(l *launcher.Launcher) {
		if bin != "" {
			l.Bin(bin)
		}
	}
}

func WithWindowSize(width, height int) LauncherOption {
	return func(l *launcher.Launcher) {
		if width > 0 && height > 0 {
			l.Set("window-size", fmt.Sprintf("%d,%d", width, height))
		}
	}
}

func WithUserAgent(ua string) LauncherOption {
	return func(l *launcher.Launcher) {
		if ua != "" {
			l.Set("user-agent", ua)
		}
	}
}

func WithDisableBackgroundNetworking(disable bool) LauncherOption {
	return func(l *launcher.Launcher) {
		l.Set("disable-background-networking", fmt.Sprintf("%v", disable))
	}
}

func WithDisableBackgroundTimerThrottling(disable bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if disable {
			l.Set("disable-background-timer-throttling")
		}
	}
}

func WithDisableBackgroundingOccludedWindows(disable bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if disable {
			l.Set("disable-backgrounding-occluded-windows")
		}
	}
}

func WithDisableRendererBackgrounding(disable bool) LauncherOption {
	return func(l *launcher.Launcher) {
		if disable {
			l.Set("disable-renderer-backgrounding")
		}
	}
}

func WithRemoteDebuggingPort(port int) LauncherOption {
	return func(l *launcher.Launcher) {
		if port > 0 {
			l.RemoteDebuggingPort(port)
		}
		if port == 0 {
			l.RemoteDebuggingPort(0)
		}
	}
}
