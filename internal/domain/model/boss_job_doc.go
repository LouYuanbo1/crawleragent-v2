package model

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/densevectorelementtype"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/densevectorsimilarity"
)

type BossJobDoc struct {
	EncryptJobId     string    `json:"encryptJobId"`
	JobName          string    `json:"jobName"`
	SalaryDesc       string    `json:"salaryDesc"`
	BrandName        string    `json:"brandName"`
	BrandScaleName   string    `json:"brandScaleName"`
	CityName         string    `json:"cityName"`
	AreaDistrict     string    `json:"areaDistrict"`
	BusinessDistrict string    `json:"businessDistrict"`
	JobLabels        []string  `json:"jobLabels"`
	Skills           []string  `json:"skills"`
	JobExperience    string    `json:"jobExperience"`
	JobDegree        string    `json:"jobDegree"`
	WelfareList      []string  `json:"welfareList"`
	DetailAddress    string    `json:"detailAddress"`
	Embedding        []float32 `json:"embedding"`
}

func (jd *BossJobDoc) GetID() string {
	return jd.EncryptJobId
}

func (jd *BossJobDoc) GetIndex() string {
	return "boss_jobs"
}

// GetTypeMapping 获取BossJobDoc的索引映射，用于创建带有词嵌入索引
// 用户需要根据实际情况自定义映射,这里只映射了Embedding向量字段
// 其他字段Elasticsearch会自动映射,无需自定义
func (jd *BossJobDoc) GetTypeMapping() *types.TypeMapping {
	dims := 768
	elementType := densevectorelementtype.Float
	similarity := densevectorsimilarity.Cosine
	index := true
	return &types.TypeMapping{
		Properties: map[string]types.Property{
			"embedding": types.DenseVectorProperty{
				Dims:        &dims,
				ElementType: &elementType,
				Similarity:  &similarity,
				Index:       &index,
				Type:        "dense_vector",
			},
		},
	}
}

func (jd *BossJobDoc) GetFieldNameVector() string {
	return "embedding"
}

// GetEmbeddingString 获取BossJobDoc的词嵌入字符串，用于生成词嵌入
func (jd *BossJobDoc) GetEmbeddingString() string {
	embeddingString := fmt.Sprintf("工作名:%s. 公司名:%s. 公司规模:%s. 城市:%s. 区域:%s. 商圈:%s. 标签:%s. 技能:%s. 经验:%s. 学历:%s. 福利:%s.",
		jd.JobName,
		jd.BrandName,
		jd.BrandScaleName,
		jd.CityName,
		jd.AreaDistrict,
		jd.BusinessDistrict,
		jd.JobLabels,
		jd.Skills,
		jd.JobExperience,
		jd.JobDegree,
		jd.WelfareList,
	)
	return embeddingString
}

func (jd *BossJobDoc) SetEmbedding(embedding []float32) {
	jd.Embedding = embedding
}

func (jd *BossJobDoc) GetEmbedding() []float32 {
	return jd.Embedding
}
