package param

import (
	"crawleragent-v2/internal/domain/model"
	"fmt"
	"time"
)

// Action 接口
type Action interface {
	Validate() error
}

type BaseParams struct {
	Delay time.Duration `json:"delay"`
}

// Click 特定参数
type ClickAction struct {
	BaseParams
	Selector string `json:"selector"`
}

func (c *ClickAction) Validate() error {
	if c.Selector == "" {
		return fmt.Errorf("点击操作必须指定选择器")
	}
	return nil
}

type ClickXAction struct {
	BaseParams
	Selector string `json:"selector"`
}

func (c *ClickXAction) Validate() error {
	if c.Selector == "" {
		return fmt.Errorf("点击操作必须指定选择器")
	}
	return nil
}

// Scroll 特定参数
type ScrollAction struct {
	BaseParams
	ScrollY int `json:"scroll_y"`
}

func (s *ScrollAction) Validate() error {
	if s.ScrollY == 0 {
		return fmt.Errorf("滚动操作必须指定滚动距离")
	}
	return nil
}

type JavaScriptAction struct {
	BaseParams
	JavaScript      string                                         `json:"javascript"`        // 要执行的 JavaScript 代码
	JavaScriptArgs  []any                                          `json:"javascript_args"`   // JavaScript 参数
	ContentChanSize int                                            `json:"content_chan_size"` // 用于传递 JavaScript 执行结果，不序列化
	ToDocFunc       func(content []byte) ([]model.Document, error) `json:"to_doc_func"`
}

func (j *JavaScriptAction) Validate() error {
	if j.JavaScript == "" {
		return fmt.Errorf("JavaScript操作必须指定JavaScript代码")
	}
	return nil
}
