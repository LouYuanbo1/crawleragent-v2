package agent

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type NavigateParams struct {
	URL            string `json:"url" jsonschema:"description=用于指定要导航的目标 URL"`
	EnableFullInfo bool   `json:"enable_full_info" jsonschema:"description=是否返回完整信息，false=只提取关键信息"` // true=返回完整信息，false=只提取关键信息
}

func NavigateFunc(ctx context.Context, crawler Crawler) func(ctx context.Context, params *NavigateParams) (string, error) {
	return func(ctx context.Context, params *NavigateParams) (string, error) {
		cleanURL := strings.TrimSpace(params.URL)
		// 移除首尾的反引号（解决LLM流式输出时带Markdown反引号的问题）
		cleanURL = strings.Trim(cleanURL, "`")
		// 再次Trim掉反引号移除后可能残留的空格
		cleanURL = strings.TrimSpace(cleanURL)
		err := crawler.Navigate(cleanURL)
		if err != nil {
			return "", err
		}
		if params.EnableFullInfo {
			return crawler.GetHTML()
		}
		return crawler.GetArticle()
	}
}

func NewNavigateTool(ctx context.Context, crawler Crawler) (tool.InvokableTool, error) {
	navigateTool, err := utils.InferTool(
		"navigate_tool", // tool name
		`Navigate the browser to a specified URL and wait for the page to fully load;
		use this to visit any web page, e.g.: https://www.example.com`, // tool description
		NavigateFunc(ctx, crawler))
	if err != nil {
		return nil, err
	}
	return navigateTool, nil
}

type ClickParams struct {
	Selector       string `json:"selector" jsonschema:"description=用于指定要点击的元素选择器"`
	IsXPath        bool   `json:"is_xpath" jsonschema:"description=是否使用XPath选择器"`
	EnableFullInfo bool   `json:"enable_full_info" jsonschema:"description=是否返回完整信息，false=只提取关键信息"` // true=返回完整信息，false=只提取关键信息
}

func ClickFunc(ctx context.Context, crawler Crawler) func(ctx context.Context, params *ClickParams) (string, error) {
	return func(ctx context.Context, params *ClickParams) (string, error) {
		var err error
		if params.IsXPath {
			err = crawler.DoClickX(params.Selector)
		} else {
			err = crawler.DoClick(params.Selector)
		}
		if err != nil {
			return "", err
		}
		if params.EnableFullInfo {
			return crawler.GetHTML()
		}
		return crawler.GetArticle()
	}
}

func NewClickTool(ctx context.Context, crawler Crawler) (tool.InvokableTool, error) {
	clickTool, err := utils.InferTool(
		"click_tool", // tool name
		`Web browser are used to click elements; 
		they can be used to click elements on the page using CSS selector or XPath;
			`, // tool description
		ClickFunc(ctx, crawler))
	if err != nil {
		return nil, err
	}
	return clickTool, nil
}

type ScrollParams struct {
	ScrollY        int  `json:"scroll_y" jsonschema:"description=用于指定要滚动的像素数"`
	EnableFullInfo bool `json:"enable_full_info" jsonschema:"description=是否返回完整信息，false=只提取关键信息"` // true=返回完整信息，false=只提取关键信息
}

func ScrollFunc(ctx context.Context, crawler Crawler) func(ctx context.Context, params *ScrollParams) (string, error) {
	return func(ctx context.Context, params *ScrollParams) (string, error) {
		err := crawler.DoScroll(params.ScrollY)
		if err != nil {
			return "", err
		}
		if params.EnableFullInfo {
			return crawler.GetHTML()
		}
		return crawler.GetArticle()
	}
}

func NewScrollTool(ctx context.Context, crawler Crawler) (tool.InvokableTool, error) {
	scrollTool, err := utils.InferTool(
		"scroll_tool", // tool name
		`Web browser are used to scroll the page; 
		they can be used to scroll the page by a specified amount of pixels;`, // tool description
		ScrollFunc(ctx, crawler))
	if err != nil {
		return nil, err
	}
	return scrollTool, nil
}

type JavaScriptParams struct {
	JS   string `json:"js" jsonschema:"description=用于指定要执行的JavaScript代码"`
	Args []any  `json:"args" jsonschema:"description=用于指定JavaScript代码的参数"`
}

func JavaScriptFunc(ctx context.Context, crawler Crawler) func(ctx context.Context, params *JavaScriptParams) (string, error) {
	return func(ctx context.Context, params *JavaScriptParams) (string, error) {
		result, err := crawler.DoJavaScript(ctx, params.JS, params.Args...)
		if err != nil {
			return "", err
		}
		return result, nil
	}
}

func NewJavaScriptTool(ctx context.Context, crawler Crawler) (tool.InvokableTool, error) {
	jsTool, err := utils.InferTool(
		"javascript_tool", // tool name
		`Web browser are used to execute JavaScript code on the page; 
		they can be used to execute JavaScript code on the page, eg: document.querySelector('button').click()`, // tool description
		JavaScriptFunc(ctx, crawler))
	if err != nil {
		return nil, err
	}
	return jsTool, nil
}

type HTMLParams struct {
	//URL            string `json:"url" jsonschema:"description=用于指定要打开的URL"`
	EnableFullInfo bool `json:"enable_full_info" jsonschema:"description=是否返回完整信息，false=只提取关键信息"` // true=返回完整信息，false=只提取关键信息
}

func HTMLFunc(ctx context.Context, crawler Crawler) func(ctx context.Context, params *HTMLParams) (string, error) {
	return func(ctx context.Context, params *HTMLParams) (string, error) {
		if params.EnableFullInfo {
			return crawler.GetHTML()
		}
		return crawler.GetArticle()
	}
}

func NewHTMLTool(ctx context.Context, crawler Crawler) (tool.InvokableTool, error) {
	htmlTool, err := utils.InferTool(
		"html_tool", // tool name
		`Web browser are used to get the HTML content of this page`, // tool description
		HTMLFunc(ctx, crawler))
	if err != nil {
		return nil, err
	}
	return htmlTool, nil
}
