package storage

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	collection *mongo.Collection
}

func NewStorage(DB *mongo.Database, collection string) *MongoDB {
	return &MongoDB{
		collection: DB.Collection(collection),
	}
}

func (d *MongoDB) CreateCredit(ctx context.Context, credit models.Credit) (string, error) {
	res, err := d.collection.InsertOne(ctx, credit)
	if err != nil {
		return "", fmt.Errorf("insert one failed:%s", err)
	}

	objectID, ok := res.InsertedID.(primitive.ObjectID) //Inserted ID type interface convert to type ObjectID
	if !ok {
		return "", fmt.Errorf("failed to get ObjectID:%s", err)
	}

	return objectID.Hex(), nil //convert type ObjectID to string
}

func (d *MongoDB) GetCredits(ctx context.Context) ([]models.Credit, error) {
	//получаем все доки в коллекции.
	res, err := d.collection.Find(ctx, bson.M{}) //в bson.M{} хранятся поля,которые мы хотим получить из коллекции
	if err != nil {
		return []models.Credit{}, fmt.Errorf("find failed:%s", err)
	}
	defer res.Close(ctx) //для избежания утечки и освобождения рес-ов

	var credits []models.Credit

	for res.Next(ctx) { //перемещаемся к след.результату в ответе
		var credit models.Credit

		if err = res.Decode(&credit); err != nil { //из запроса переносим данные на структуру
			return nil, fmt.Errorf("decode failed:%s", err)
		}
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return []models.Credit{}, fmt.Errorf("no credits found with provided ID:%s", credit.ID)
		}

		credits = append(credits, credit)
	}

	if res.Err() != nil {
		return nil, fmt.Errorf("failed to find credits:%s", err.Error())
	}

	return credits, nil
}

func (d *MongoDB) GetCreditById(ctx context.Context, id string) (credit models.Credit, err error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return credit, fmt.Errorf("failed to conver string to ObjectID")
	}

	query := bson.M{"_id": objectID} //ObjectID используется в кач.значения поля _id

	res := d.collection.FindOne(ctx, query)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return credit, fmt.Errorf("no credit found with provided ID:%s", id)
	}
	if res.Err() != nil {
		return credit, fmt.Errorf("failed to find credit by id:%s", err.Error())
	}

	if err = res.Decode(&credit); err != nil {
		return credit, fmt.Errorf("decode failed:%s", err)
	}

	return credit, nil
}

func (d *MongoDB) UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error) {
	objectID, err := primitive.ObjectIDFromHex(credit.ID)
	if err != nil {
		return updatedCredit, fmt.Errorf("failed to conver string to ObjectID")
	}

	query := bson.M{"_id": objectID}

	update := bson.M{
		"$set": bson.M{ //поля,которые нужно обновить
			"amount":         credit.Amount,
			"dateOfIssue":    credit.DateOfIssue,
			"maturityDate":   credit.MaturityDate,
			"term":           credit.Term,
			"monthlyPayment": credit.MonthlyPayment,
		},
	}

	res := d.collection.FindOneAndUpdate(ctx, query, update, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return updatedCredit, fmt.Errorf("no credit found with provided ID")
	}
	if res.Err() != nil {
		return models.Credit{}, fmt.Errorf("failed to update credit:%s", err)
	}

	if err = res.Decode(&updatedCredit); err != nil {
		return updatedCredit, fmt.Errorf("decode failed:%s", err)
	}

	return updatedCredit, nil
}

func (d *MongoDB) DeleteCredit(ctx context.Context, id string) error {
	ObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to conver string to ObjectID")
	}

	query := bson.M{"_id": ObjectID}

	res, err := d.collection.DeleteOne(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete credit:%s", err)
	}

	if res.DeletedCount == 0 {
		return fmt.Errorf("no credit found with provided ID:%s", id)
	}

	return nil
}
