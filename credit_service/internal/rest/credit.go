package rest

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
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

		h.IsNilBodyReq(w, r)

		var credit models.Credit

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		objectId, err := h.service.CreateCredit(context.Background(), credit)
		if err != nil {
			h.logger.Errorf("create credit failed:%s", err)
			http.Error(w, "create credit failed", http.StatusInternalServerError)
			return
		}

		response := map[string]string{"ObjectId": objectId}

		render.JSON(w, r, response)
	}
}

func (h *Handler) GetCredits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		h.IsNilBodyReq(w, r)

		credits, err := h.service.GetCredits(context.Background())
		if err != nil {
			h.logger.Errorf("get credits failed:%s", err)
			http.Error(w, "get credits failed", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, credits)
	}
}

func (h *Handler) GetCreditById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		h.IsNilBodyReq(w, r)

		creditID := chi.URLParam(r, "id") //получаем id из url req

		credit, err := h.service.GetCreditById(context.Background(), creditID)
		if err != nil {
			h.logger.Errorf("get credit by id failed:%s", err)
			http.Error(w, "get credit by id failed", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, credit)
	}
}

func (h *Handler) UpdateCredit() http.HandlerFunc { //return credit
	return func(w http.ResponseWriter, r *http.Request) {

		h.IsNilBodyReq(w, r)

		var credit models.Credit

		credit.ID = chi.URLParam(r, "id")

		if err := h.decodeJSONFromBody(w, r, &credit); err != nil {
			return
		}

		updatedCredit, err := h.service.UpdateCredit(context.Background(), credit)
		if err != nil {
			h.logger.Errorf("update credit failed:%s", err)
			http.Error(w, "update credit failed", http.StatusInternalServerError)
			return
		}

		response := map[string]models.Credit{"Updated Credit": updatedCredit}

		render.JSON(w, r, response)
	}
}

func (h *Handler) DeleteCredit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		h.IsNilBodyReq(w, r)

		creditID := chi.URLParam(r, "id")

		if err := h.service.DeleteCredit(context.Background(), creditID); err != nil {
			h.logger.Errorf("delete credit failed:%s", err)
			http.Error(w, "delete credit failed", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, "deleted successfully")
	}
}

func (h *Handler) IsNilBodyReq(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Errorf("request body empty")
		http.Error(w, "request body empty", http.StatusBadRequest)
		return
	}
}

func (h *Handler) decodeJSONFromBody(w http.ResponseWriter, r *http.Request, data interface{}) error {
	if err := render.DecodeJSON(r.Body, data); err != nil {
		h.logger.Errorf("failed to decode request body: %s", err)
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return err
	}
	return nil
}
