package controller

type QueryWithPrompt struct {
	Index           string `json:"index,omitempty"`
	Query           string `json:"query,omitempty"`
	PromptEsRAGMode string `json:"promptEsRAGMode,omitempty"`
	PromptChatMode  string `json:"promptChatMode,omitempty"`
}
