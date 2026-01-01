package embedding

import (
	"context"
)

// Embedder 嵌入器接口,用于将文本转换为向量表示
type Embedder interface {
	Embed(ctx context.Context, strings []string) ([][]float32, error)
}
