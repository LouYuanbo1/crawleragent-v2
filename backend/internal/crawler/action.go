package crawler

import (
	"context"
	"fmt"
	"time"
)

// Action 页面操作接口
type Action interface {
	Validate() error
}

// ActionBase 所有操作的公共参数
type ActionBase struct {
	Delay time.Duration `json:"delay"`
}

// ClickAction CSS选择器点击
type ClickAction struct {
	ActionBase
	Selector string `json:"selector"`
}

func (a *ClickAction) Validate() error {
	if a.Selector == "" {
		return fmt.Errorf("点击操作必须指定选择器")
	}
	return nil
}

// ClickXAction XPath选择器点击
type ClickXAction struct {
	ActionBase
	Selector string `json:"selector"`
}

func (a *ClickXAction) Validate() error {
	if a.Selector == "" {
		return fmt.Errorf("点击操作必须指定选择器")
	}
	return nil
}

// ScrollAction 页面滚动
type ScrollAction struct {
	ActionBase
	ScrollY int `json:"scroll_y"`
}

func (a *ScrollAction) Validate() error {
	if a.ScrollY == 0 {
		return fmt.Errorf("滚动操作必须指定滚动距离")
	}
	return nil
}

// JavaScriptAction 执行JS并处理结果
type JavaScriptAction struct {
	ActionBase
	JavaScript     string `json:"javascript"`
	JavaScriptArgs []any  `json:"javascript_args"`
	ProcessFunc    func(ctx context.Context, content UrlContent) error
}

func (a *JavaScriptAction) Validate() error {
	if a.JavaScript == "" {
		return fmt.Errorf("JavaScript操作必须指定JavaScript代码")
	}
	return nil
}
