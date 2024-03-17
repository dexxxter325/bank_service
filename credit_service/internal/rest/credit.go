package rest

//TODO: узнать ,зачем возвращаем return func ,интерфейсы по месту использования,раскатка в кубер
import (
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

	}

}

func (h *Handler) GetCredits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) GetCreditById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) UpdateCredit() http.HandlerFunc { //return credit
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) DeleteCredit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
