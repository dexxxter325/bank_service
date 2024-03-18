package models

import "time"

type Credit struct {
	ID             string    `bson:"_id,omitempty"`
	Amount         int       `bson:"amount"`
	DateOfIssue    time.Time `bson:"dateOfIssue"`
	MaturityDate   time.Time `bson:"maturityDate"`
	Term           int       `bson:"term"`
	MonthlyPayment int       `bson:"monthlyPayment"`
}
