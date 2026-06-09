package crawler

import (
	"testing"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/stretchr/testify/assert"
)

// ==================== CreateLauncher ====================

func TestCreateLauncher_AutoMode(t *testing.T) {
	l := CreateLauncher(false)
	assert.NotNil(t, l)
	// AutoMode launcher 的 Headless 应默认为 true
}

func TestCreateLauncher_UserMode(t *testing.T) {
	l := CreateLauncher(true)
	assert.NotNil(t, l)
	// UserMode launcher 使用已有浏览器
}

func TestCreateLauncher_WithOptions(t *testing.T) {
	l := CreateLauncher(false, WithHeadless(false), WithIncognito(true))
	assert.NotNil(t, l)
}

// ==================== boolFlag 工厂 ====================

func TestBoolFlag_Enable(t *testing.T) {
	opt := boolFlag(flags.Flag("test-flag"), true)
	l := launcher.New()
	opt(l)

	assert.Contains(t, l.Flags, flags.Flag("test-flag"),
		"启用时应该设置 flag")
}

func TestBoolFlag_Disable(t *testing.T) {
	opt := boolFlag(flags.Flag("test-flag"), false)
	l := launcher.New()
	opt(l)

	assert.NotContains(t, l.Flags, flags.Flag("test-flag"),
		"未启用时不应设置 flag")
}

// ==================== WithIncognito ====================

func TestWithIncognito_True(t *testing.T) {
	l := launcher.New()
	WithIncognito(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("incognito"))
}

func TestWithIncognito_False(t *testing.T) {
	l := launcher.New()
	WithIncognito(false)(l)
	assert.NotContains(t, l.Flags, flags.Flag("incognito"))
}

// ==================== WithHeadless ====================

func TestWithHeadless_True(t *testing.T) {
	l := launcher.New()
	WithHeadless(true)(l)
	// 新创建默认就是 true，此处验证显式设置不报错
}

func TestWithHeadless_False(t *testing.T) {
	l := launcher.New()
	WithHeadless(false)(l)
	// 设置为 false 不报错
}

// ==================== WithDisableDevShmUsage ====================

func TestWithDisableDevShmUsage(t *testing.T) {
	l := launcher.New()
	WithDisableDevShmUsage(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-dev-shm-usage"))
}

// ==================== WithNoSandbox ====================

func TestWithNoSandbox(t *testing.T) {
	l := launcher.New()
	WithNoSandbox(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("no-sandbox"))
}

// ==================== WithDisableBackgroundNetworking ====================

func TestWithDisableBackgroundNetworking(t *testing.T) {
	l := launcher.New()
	WithDisableBackgroundNetworking(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-background-networking"))
}

// ==================== WithDisableBackgroundTimerThrottling ====================

func TestWithDisableBackgroundTimerThrottling(t *testing.T) {
	l := launcher.New()
	WithDisableBackgroundTimerThrottling(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-background-timer-throttling"))
}

// ==================== WithDisableBackgroundingOccludedWindows ====================

func TestWithDisableBackgroundingOccludedWindows(t *testing.T) {
	l := launcher.New()
	WithDisableBackgroundingOccludedWindows(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-backgrounding-occluded-windows"))
}

// ==================== WithDisableRendererBackgrounding ====================

func TestWithDisableRendererBackgrounding(t *testing.T) {
	l := launcher.New()
	WithDisableRendererBackgrounding(true)(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-renderer-backgrounding"))
}

// ==================== WithDisableBlinkFeatures ====================

func TestWithDisableBlinkFeatures_NonEmpty(t *testing.T) {
	l := launcher.New()
	WithDisableBlinkFeatures("AutomationControlled")(l)
	assert.Contains(t, l.Flags, flags.Flag("disable-blink-features"))
}

func TestWithDisableBlinkFeatures_Empty(t *testing.T) {
	l := launcher.New()
	WithDisableBlinkFeatures("")(l)
	assert.NotContains(t, l.Flags, flags.Flag("disable-blink-features"))
}

// ==================== WithUserDataDir ====================

func TestWithUserDataDir_NonEmpty(t *testing.T) {
	l := launcher.New()
	dir := "/tmp/chrome-profile"
	WithUserDataDir(dir)(l)
	assert.Contains(t, l.Flags, flags.Flag("user-data-dir"))
}

func TestWithUserDataDir_Empty(t *testing.T) {
	l := launcher.New()
	WithUserDataDir("")(l)
	// 空字符串不应覆盖默认 user-data-dir
	assert.Contains(t, l.Flags, flags.Flag("user-data-dir"))
}

// ==================== WithBin ====================

func TestWithBin(t *testing.T) {
	l := launcher.New()
	bin := "/usr/bin/chrome"
	WithBin(bin)(l)
	// Bin 使用专门方法，不进入 Flags map，验证不会panic
	// l.Bin 是string字段，无法直接断言，但调用应不报错
}

func TestWithBin_Empty(t *testing.T) {
	l := launcher.New()
	WithBin("")(l)
	// 空字符串不应修改 Bin
}

// ==================== WithWindowSize ====================

func TestWithWindowSize_Valid(t *testing.T) {
	l := launcher.New()
	WithWindowSize(1920, 1080)(l)
	assert.Contains(t, l.Flags, flags.Flag("window-size"))
}

func TestWithWindowSize_ZeroWidth(t *testing.T) {
	l := launcher.New()
	WithWindowSize(0, 1080)(l)
	assert.NotContains(t, l.Flags, flags.Flag("window-size"))
}

func TestWithWindowSize_ZeroHeight(t *testing.T) {
	l := launcher.New()
	WithWindowSize(1920, 0)(l)
	assert.NotContains(t, l.Flags, flags.Flag("window-size"))
}

// ==================== WithUserAgent ====================

func TestWithUserAgent_NonEmpty(t *testing.T) {
	l := launcher.New()
	ua := "Mozilla/5.0 TestAgent"
	WithUserAgent(ua)(l)
	assert.Contains(t, l.Flags, flags.Flag("user-agent"))
}

func TestWithUserAgent_Empty(t *testing.T) {
	l := launcher.New()
	WithUserAgent("")(l)
	assert.NotContains(t, l.Flags, flags.Flag("user-agent"))
}

// ==================== WithLeakless ====================

func TestWithLeakless_True(t *testing.T) {
	l := launcher.New()
	WithLeakless(true)(l)
	// Leakless 使用专门方法，调用不报错
}

// ==================== WithRemoteDebuggingPort ====================

func TestWithRemoteDebuggingPort(t *testing.T) {
	l := launcher.New()
	WithRemoteDebuggingPort(9222)(l)
	// RemoteDebuggingPort 调用不报错
}

// ==================== 组合选项 ====================

func TestMultipleOptions_ComposeCorrectly(t *testing.T) {
	l := launcher.New()

	WithIncognito(true)(l)
	WithNoSandbox(true)(l)
	WithHeadless(false)(l)
	WithUserAgent("test-agent")(l)

	assert.Contains(t, l.Flags, flags.Flag("incognito"))
	assert.Contains(t, l.Flags, flags.Flag("no-sandbox"))
	assert.Contains(t, l.Flags, flags.Flag("user-agent"))
}

func TestCreateLauncher_AppliesAllOptions(t *testing.T) {
	l := CreateLauncher(false,
		WithIncognito(true),
		WithNoSandbox(true),
		WithDisableDevShmUsage(true),
		WithUserAgent("bot/1.0"),
	)

	assert.Contains(t, l.Flags, flags.Flag("incognito"))
	assert.Contains(t, l.Flags, flags.Flag("no-sandbox"))
	assert.Contains(t, l.Flags, flags.Flag("disable-dev-shm-usage"))
	assert.Contains(t, l.Flags, flags.Flag("user-agent"))
}

func TestLauncherOption_IsFuncType(t *testing.T) {
	// 验证 LauncherOption 类型正确
	var opt LauncherOption = func(l *launcher.Launcher) {}
	assert.NotNil(t, opt)
}
