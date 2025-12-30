package embedding

import (
	"context"
	"crawleragent-v2/internal/config"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ollama"
	"golang.org/x/sync/semaphore"
)

type embedding struct {
	model     *ollama.Embedder
	batchSize int
	embedSem  *semaphore.Weighted
}

// InitEmbedder 初始化嵌入器
func InitEmbedder(ctx context.Context, cfg *config.Config, batchSize int, embedSemSize int) (Embedder, error) {
	model, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
		Model:   cfg.Embedding.Model,
		BaseURL: cfg.Embedding.Host + ":" + strconv.Itoa(cfg.Embedding.Port),
	})
	if err != nil {
		return nil, err
	}
	embedSem := semaphore.NewWeighted(int64(embedSemSize))
	return &embedding{model: model, batchSize: batchSize, embedSem: embedSem}, nil
}

// Embed 将文本转换为向量表示
func (e *embedding) Embed(ctx context.Context, strings []string) ([][]float32, error) {
	if len(strings) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	embeddingVectors := make([][]float32, 0, len(strings))
	var err error

	log.Printf("嵌入 %d 个字符串", len(strings))

	if len(strings) <= e.batchSize {
		embeddingVectors, err = e.batchEmbedStringsAndToFloat32(ctx, strings)
		if err != nil {
			return nil, err
		}
	} else {
		for i := 0; i < len(strings); i += e.batchSize {
			end := min(i+e.batchSize, len(strings))
			batch := strings[i:end]
			batchEmbeddingVectors, err := e.batchEmbedStringsAndToFloat32(ctx, batch)
			if err != nil {
				return nil, err
			}
			embeddingVectors = append(embeddingVectors, batchEmbeddingVectors...)
		}
	}
	//log.Printf("嵌入 %d 个字符串，得到 %d 个向量", len(strings), len(embeddingVectors))
	return embeddingVectors, nil
}

func (e *embedding) batchEmbedStringsAndToFloat32(ctx context.Context, strings []string) ([][]float32, error) {
	// 获取信号量（带超时）
	if err := e.embedSem.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("等待词嵌入信号量超时: %w", err)
	}
	defer e.embedSem.Release(1) // 保证释放

	float64Vectors, err := e.model.EmbedStrings(ctx, strings)
	if err != nil {
		return nil, err
	}
	//log.Printf("转换 %d 个 float64 向量为 float32 向量", len(float64Vectors))
	// 外层二维切片：预分配容量，避免频繁扩容
	float32Vectors := make([][]float32, 0, len(float64Vectors))

	for _, float64Vector := range float64Vectors {
		// 内层一维切片：变量名换为 innerFloat32Vec，避免覆盖外层
		innerFloat32Vec := make([]float32, 0, len(float64Vector))

		// 逐个元素转换 float64 -> float32
		for _, f := range float64Vector {
			innerFloat32Vec = append(innerFloat32Vec, float32(f))
		}

		// 关键修复：将内层一维切片追加到外层二维切片中
		float32Vectors = append(float32Vectors, innerFloat32Vec)
	}

	return float32Vectors, nil
}
