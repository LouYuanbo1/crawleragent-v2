package entity

import (
	"crawleragent-v2/internal/domain/model"
	"fmt"
)

// RowBossJobData是Boss直聘上的原始数据
type RowBossJobData struct {
	EncryptJobId     string   `json:"encryptJobId"`
	SecurityId       string   `json:"securityId"`
	JobName          string   `json:"jobName"`
	SalaryDesc       string   `json:"salaryDesc"`
	BrandName        string   `json:"brandName"`
	BrandScaleName   string   `json:"brandScaleName"`
	CityName         string   `json:"cityName"`
	AreaDistrict     string   `json:"areaDistrict"`
	BusinessDistrict string   `json:"businessDistrict"`
	JobLabels        []string `json:"jobLabels"`
	Skills           []string `json:"skills"`
	JobExperience    string   `json:"jobExperience"`
	JobDegree        string   `json:"jobDegree"`
	WelfareList      []string `json:"welfareList"`
}

// ToDocument 将RowBossJobData转换为BossJobDoc
// Document类型用于将原始数据转换为索引文档(保存到es)
func (entity *RowBossJobData) ToDocument() model.Document {
	return &model.BossJobDoc{
		EncryptJobId:     entity.EncryptJobId,
		JobName:          entity.JobName,
		SalaryDesc:       entity.SalaryDesc,
		BrandName:        entity.BrandName,
		BrandScaleName:   entity.BrandScaleName,
		CityName:         entity.CityName,
		AreaDistrict:     entity.AreaDistrict,
		BusinessDistrict: entity.BusinessDistrict,
		JobLabels:        entity.JobLabels,
		Skills:           entity.Skills,
		JobExperience:    entity.JobExperience,
		JobDegree:        entity.JobDegree,
		WelfareList:      entity.WelfareList,
		DetailAddress: fmt.Sprintf("https://www.zhipin.com/job_detail/%s.html?securityId=%s&ka=company_more_job_%s",
			entity.EncryptJobId, entity.SecurityId, entity.EncryptJobId),
	}
}

// 将来可能爬bilibili,先这样吧
type RowBiliVideoData struct {
	Name           string `json:"name"`
	Bvid           string `json:"bvid"`
	Views          int64  `json:"views"`
	Likes          int64  `json:"likes"`
	Coins          int64  `json:"coins"`
	Favorites      int64  `json:"favorites"`
	Comments       int64  `json:"comments"`
	BulletComments int64  `json:"bulletComments"`
}
