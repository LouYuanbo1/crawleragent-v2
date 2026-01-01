package es

import (
	"context"

	"crawleragent-v2/internal/data/model"
)

/*
// 所有的文档结构体要实现这两个函数

		type Document interface {
		GetID() string
		GetIndex() string
		GetTypeMapping() *types.TypeMapping
		GetEmbeddingString() string
		SetEmbedding(embedding []float32)
		GetEmbedding() []float32
	}
*/
type TypedEsClient interface {
	CreateIndexWithMapping(ctx context.Context, doc model.Document) error
	DeleteIndex(ctx context.Context, index string) error
	IndexDocWithID(ctx context.Context, doc model.Document) error
	BulkIndexDocsWithID(ctx context.Context, docs []model.Document) error
	SearchStrDocsByVector(ctx context.Context, doc model.Document, queryVector []float32, k, numCandidates int) (string, error)
	GetDoc(ctx context.Context, index string, id string) (model.Document, error)
	CountDocs(ctx context.Context, index string) (int64, error)
	UpdateDoc(ctx context.Context, doc model.Document) error
	DeleteDoc(ctx context.Context, index string, id string) error
	BulkDeleteDocs(ctx context.Context, index string, ids []string) error
	ToExcel(ctx context.Context, filename string, index string, sortFields []string, size int) error
}
