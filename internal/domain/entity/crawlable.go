package entity

import "crawleragent-v2/internal/domain/model"

type Crawlable interface {
	ToDocument() model.Document
}
