package rest

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type Handler struct {
	logger  *logrus.Logger
	service Service
}

func NewHandler(logger *logrus.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) CreateCredit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var credit models.Credit

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		objectId, err := h.service.CreateCredit(context.Background(), credit)
		if err != nil {
			h.logger.Errorf("create credit failed:%s", err)
			http.Error(w, fmt.Sprintf("create credit failed:%s", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{"ObjectId": objectId}

		render.JSON(w, r, response)
	}
}

func (h *Handler) GetCredits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		credits, err := h.service.GetCredits(context.Background())
		if err != nil {
			h.logger.Errorf("get credits failed:%s", err)
			if errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, fmt.Sprintf("no credits found:%s", err.Error()), http.StatusNotFound)
			}
			http.Error(w, fmt.Sprintf("get credits failed:%s", err), http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, credits)
	}
}

func (h *Handler) GetCreditById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		creditID := chi.URLParam(r, "id") //получаем id из url req

		credit, err := h.service.GetCreditById(context.Background(), creditID)
		if err != nil {
			h.logger.Errorf("get credit by id failed:%s", err)
			http.Error(w, fmt.Sprintf("get credit by id failed:%s", err), http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, credit)
	}
}

func (h *Handler) UpdateCredit() http.HandlerFunc { //return credit
	return func(w http.ResponseWriter, r *http.Request) {

		var credit models.Credit

		credit.ID = chi.URLParam(r, "id")

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		updatedCredit, err := h.service.UpdateCredit(context.Background(), credit)
		if err != nil {
			h.logger.Errorf("update credit failed:%s", err)
			http.Error(w, fmt.Sprintf("update credit failed:%s", err), http.StatusInternalServerError)
			return
		}

		response := map[string]models.Credit{"Updated Credit": updatedCredit}

		render.JSON(w, r, response)
	}
}

func (h *Handler) DeleteCredit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		creditID := chi.URLParam(r, "id")

		if err := h.service.DeleteCredit(context.Background(), creditID); err != nil {
			h.logger.Errorf("delete credit failed:%s", err)
			http.Error(w, fmt.Sprintf("delete credit failed:%s", err), http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, "deleted successfully")
	}
}

func (h *Handler) decodeJSONFromBody(w http.ResponseWriter, r *http.Request, data interface{}) error {
	if err := render.DecodeJSON(r.Body, data); err != nil {
		if errors.Is(err, io.EOF) {
			h.logger.Errorf("request body is empty,you must enter the data: %s", err)
			http.Error(w, "request body is empty,you must enter the data", http.StatusBadRequest)
			return err
		}
		h.logger.Errorf("failed to decode request body: %s", err)
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return err
	}
	return nil
}
