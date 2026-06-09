package crawler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== ClickAction ====================

func TestClickAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		action  *ClickAction
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效选择器",
			action:  &ClickAction{Selector: "#app"},
			wantErr: false,
		},
		{
			name:    "空选择器",
			action:  &ClickAction{Selector: ""},
			wantErr: true,
			errMsg:  "必须指定选择器",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClickAction_ImplementsAction(t *testing.T) {
	var a Action = &ClickAction{Selector: "#app"}
	assert.Implements(t, (*Action)(nil), a)
}

func TestClickAction_ActionBaseDelay(t *testing.T) {
	a := &ClickAction{ActionBase: ActionBase{Delay: 100}}
	assert.Equal(t, time.Duration(100), a.Delay)
}

// ==================== ClickXAction ====================

func TestClickXAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		action  *ClickXAction
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效XPath",
			action:  &ClickXAction{Selector: "//div[@id='app']"},
			wantErr: false,
		},
		{
			name:    "空XPath",
			action:  &ClickXAction{Selector: ""},
			wantErr: true,
			errMsg:  "必须指定选择器",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClickXAction_ImplementsAction(t *testing.T) {
	var a Action = &ClickXAction{Selector: "//div"}
	assert.Implements(t, (*Action)(nil), a)
}

// ==================== ScrollAction ====================

func TestScrollAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		action  *ScrollAction
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效滚动",
			action:  &ScrollAction{ScrollY: 300},
			wantErr: false,
		},
		{
			name:    "负值滚动",
			action:  &ScrollAction{ScrollY: -100},
			wantErr: false,
		},
		{
			name:    "零值滚动",
			action:  &ScrollAction{ScrollY: 0},
			wantErr: true,
			errMsg:  "必须指定滚动距离",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScrollAction_ImplementsAction(t *testing.T) {
	var a Action = &ScrollAction{ScrollY: 100}
	assert.Implements(t, (*Action)(nil), a)
}

// ==================== JavaScriptAction ====================

func TestJavaScriptAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		action  *JavaScriptAction
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效JS",
			action:  &JavaScriptAction{JavaScript: "document.title"},
			wantErr: false,
		},
		{
			name:    "空JS",
			action:  &JavaScriptAction{JavaScript: ""},
			wantErr: true,
			errMsg:  "必须指定JavaScript代码",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJavaScriptAction_ImplementsAction(t *testing.T) {
	var a Action = &JavaScriptAction{JavaScript: "1+1"}
	assert.Implements(t, (*Action)(nil), a)
}

func TestJavaScriptAction_WithProcessFunc(t *testing.T) {
	called := false
	a := &JavaScriptAction{
		JavaScript: "document.title",
		ProcessFunc: func(ctx context.Context, content UrlContent) error {
			called = true
			return nil
		},
	}
	assert.NoError(t, a.Validate())
	assert.NotNil(t, a.ProcessFunc)

	// 验证ProcessFunc可被调用
	_ = a.ProcessFunc(context.Background(), &HtmlContent{Url: "http://example.com", Content: []byte("test")})
	assert.True(t, called)
}

func TestJavaScriptAction_WithArgs(t *testing.T) {
	a := &JavaScriptAction{
		JavaScript:     "(x, y) => x + y",
		JavaScriptArgs: []any{1, 2},
	}
	assert.NoError(t, a.Validate())
	assert.Equal(t, []any{1, 2}, a.JavaScriptArgs)
}

// ==================== 多个 Action 作为接口使用 ====================

func TestActionSlice(t *testing.T) {
	actions := []Action{
		&ClickAction{Selector: "#btn"},
		&ClickXAction{Selector: "//button"},
		&ScrollAction{ScrollY: 500},
		&JavaScriptAction{JavaScript: "window.scrollTo(0, 0)"},
	}

	for i, a := range actions {
		assert.NoErrorf(t, a.Validate(), "action[%d] 验证失败", i)
	}
}
