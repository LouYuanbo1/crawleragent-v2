package model

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

type Document interface {
	GetID() string
	GetIndex() string
	GetTypeMapping() *types.TypeMapping
	GetFieldNameVector() string
	GetEmbeddingString() string
	SetEmbedding(embedding []float32)
	GetEmbedding() []float32
}

func UnmarshalDocument(index string, data []byte) (Document, error) {
	switch index {
	case "boss_jobs":
		var doc BossJobDoc
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		return &doc, nil
	default:
		return nil, fmt.Errorf("unknown document type: %s", index)
	}
}

func IndexToDoc(index string) (Document, error) {
	if index == "boss_jobs" {
		return &BossJobDoc{}, nil
	}
	return nil, fmt.Errorf("index %s not supported", index)
}
