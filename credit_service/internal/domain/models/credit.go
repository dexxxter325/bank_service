package models

type Credit struct {
	ID                 string  `bson:"_id,omitempty"`
	UserID             int64   `bson:"userID" validate:"required_if=OperationType create"`
	Amount             int     `bson:"amount" validate:"required"`
	Currency           string  `bson:"currency" validate:"required"`
	AnnualInterestRate float64 `bson:"annualInterestRate" validate:"required"` //годовая % ставка
	Term               int     `bson:"term" validate:"required"`               //срок кредита
	DateOfIssue        string  `bson:"dateOfIssue"`                            //дата выдачи
	MaturityDate       string  `bson:"maturityDate"`                           //срок погашения
	MonthlyPayment     int     `bson:"monthlyPayment"`
	OperationType      string
}
