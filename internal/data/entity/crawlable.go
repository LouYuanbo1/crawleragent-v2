package entity

import "crawleragent-v2/internal/data/model"

type Crawlable interface {
	ToDocument() model.Document
}
