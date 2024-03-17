package rest

import (
	"bank/credit_service/internal/domain/models"
	"github.com/go-chi/chi"
	"net/http"
)

type Service interface {
	CreateCredit(credit *models.Credit) (int, error)
	GetCredits() (*models.Credit, error)
	GetCreditById(int) (*models.Credit, error)
	UpdateCredit(int) (*models.Credit, error)
	DeleteCredit(int) error
}

func (h *Handler) InitRoutes(r *chi.Mux) http.Handler {
	r.Route("/credits", func(r chi.Router) {
		r.Post("/", h.CreateCredit())
		r.Get("/", h.GetCredits())
		r.Get("/{id}", h.GetCreditById())
		r.Put("/{id}", h.UpdateCredit())
		r.Delete("/{id}", h.DeleteCredit())
	})
	return r
}
