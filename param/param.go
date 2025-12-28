package param

import (
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
