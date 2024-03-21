package models

type Credit struct {
	ID                 string  `bson:"_id,omitempty"`
	Amount             int     `bson:"amount"`
	Currency           string  `bson:"currency"`
	AnnualInterestRate float64 `bson:"annualInterestRate"` //годовая % ставка
	Term               int     `bson:"term"`               //срок кредита
	DateOfIssue        string  `bson:"dateOfIssue"`        //дата выдачи
	MaturityDate       string  `bson:"maturityDate"`       //срок погашения
	MonthlyPayment     int     `bson:"monthlyPayment"`
}
