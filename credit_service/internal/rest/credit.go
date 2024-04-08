package rest

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
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

		credit.OperationType = "create"
		if err := h.ValidateValues(w, &credit); err != nil {
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

func (h *Handler) GetCreditsByUserId() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userIdStr := chi.URLParam(r, "id") //получаем id из url req

		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			h.logger.Errorf("convert userId to int failed:%s", err)
			http.Error(w, fmt.Sprintf("convert userId to int failed:%s", err), http.StatusInternalServerError)
			return
		}

		userIdInt64 := int64(userId)

		credits, err := h.service.GetCreditsByUserId(context.Background(), userIdInt64)
		if err != nil {
			if strings.Contains(err.Error(), "no credits found for provided userID") {
				h.logger.Errorf("no credits found for provided userID: %v", userIdInt64)
				http.Error(w, fmt.Sprintf("no credits found for provided userID: %v", userIdInt64), http.StatusNotFound)
				return
			}
			h.logger.Errorf("get credit by userId failed:%s", err)
			http.Error(w, fmt.Sprintf("get credit by userId failed:%s", err), http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, credits)
	}
}

func (h *Handler) UpdateCredit() http.HandlerFunc { //return credit
	return func(w http.ResponseWriter, r *http.Request) {

		var credit models.Credit

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		if err := h.ValidateValues(w, &credit); err != nil {
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
		http.Error(w, fmt.Sprintf("failed to decode request body:%s", err), http.StatusBadRequest)
		return err
	}
	return nil
}

func (h *Handler) ValidateValues(w http.ResponseWriter, credit *models.Credit) error {
	validate := validator.New()

	if err := validate.Struct(credit); err != nil {
		validateErr := err.(validator.ValidationErrors)
		h.logger.Errorf("invalid req:%s", ValidationErrors(validateErr))
		http.Error(w, fmt.Sprintf("invalid req:%s", ValidationErrors(validateErr)), http.StatusBadRequest)
		return err
	}

	return nil
}
