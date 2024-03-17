package models

import "time"

type Credit struct {
	ID             int
	Amount         int
	DateOfIssue    time.Time
	MaturityDate   time.Time
	Term           int //срок кредита
	MonthlyPayment int
}
