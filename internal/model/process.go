package model

type TransactionInfo struct {
	ActiveTime int64
	Info       []string
}

type ProcessWithTransaction struct {
	ID          int64
	User        string
	Host        string
	DB          string
	Command     string
	Time        int64
	State       string
	Info        string
	Progress    float64
	Transaction *TransactionInfo
}
