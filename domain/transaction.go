package domain

type TransactionRepository interface {
	Search(reportID string, accountID string, initDate string, endDate string) error
}
