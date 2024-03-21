package rest

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"github.com/go-chi/chi"
	"net/http"
)

type Service interface {
	CreateCredit(ctx context.Context, credit models.Credit) (models.Credit, error)
	GetCredits(ctx context.Context) ([]models.Credit, error)
	GetCreditById(ctx context.Context, id string) (models.Credit, error)
	UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error)
	DeleteCredit(ctx context.Context, id string) error
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
