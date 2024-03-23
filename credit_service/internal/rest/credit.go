package rest

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
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

		if err := h.AreValuesFilled(credit.Amount, credit.Term, credit.Currency, credit.AnnualInterestRate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		createdCredit, err := h.service.CreateCredit(context.Background(), credit)
		if err != nil {
			h.logger.Errorf("create credit failed:%s", err)
			http.Error(w, fmt.Sprintf("create credit failed:%s", err), http.StatusInternalServerError)
			return
		}

		response := map[string]models.Credit{"Created Credit": createdCredit}

		render.JSON(w, r, response)
	}
}

func (h *Handler) GetCredits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		credits, err := h.service.GetCredits(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "no credits found") {
				h.logger.Error("no credits found")
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			h.logger.Errorf("get credits failed:%s", err)
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
			if strings.Contains(err.Error(), "no credit found with provided ID") {
				h.logger.Errorf("no credit found with provided ID: %s", creditID)
				http.Error(w, fmt.Sprintf("no credit found with provided ID: %s", creditID), http.StatusNotFound)
				return
			}
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

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		if err := h.AreValuesFilled(credit.Amount, credit.Term, credit.Currency, credit.AnnualInterestRate); err != nil {
			h.logger.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		credit.ID = chi.URLParam(r, "id")

		updatedCredit, err := h.service.UpdateCredit(context.Background(), credit)
		if err != nil {
			if strings.Contains(err.Error(), "no credit found with provided ID") {
				h.logger.Errorf("no credit found with provided ID: %s", credit.ID)
				http.Error(w, fmt.Sprintf("no credit found with provided ID: %s", credit.ID), http.StatusNotFound)
				return
			}
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
			if strings.Contains(err.Error(), "no credit found with provided ID") {
				h.logger.Errorf("no credit found with provided ID: %s", creditID)
				http.Error(w, fmt.Sprintf("no credit found with provided ID: %s", creditID), http.StatusNotFound)
				return
			}
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

func (h *Handler) AreValuesFilled(amount, term int, currency string, annualInterestRate float64) error {
	if amount == 0 {
		h.logger.Error("you must fill the 'amount' value")
		return errors.New("you must fill the 'amount' value")
	}
	if currency == "" {
		h.logger.Error("you must fill the 'currency' value")
		return errors.New("you must fill the 'currency' value")
	}
	if term == 0 {
		h.logger.Error("you must fill the 'term' value")
		return errors.New("you must fill the 'term' value")
	}
	if annualInterestRate == 0 {
		h.logger.Error("you must fill the 'annualInterestRate' value")
		return errors.New("you must fill the 'annualInterestRate' value")
	}
	return nil
}
