package repository

import (
	"github.com/Vractos/go-reports-service/dto"
	"github.com/elastic/go-elasticsearch/v8"
)

type TransactionElasticRepository struct {
	Client elasticsearch.Client
}

func (t TransactionElasticRepository) Search(reportID string, accountID string, initDate string, endDate string) (dto.SearchResponse, error) {

}
