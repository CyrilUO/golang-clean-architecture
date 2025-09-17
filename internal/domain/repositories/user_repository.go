package repositories

import (
	"clean-archi-analytics/internal/domain/entities"
	"context"
)

// UserRepository définit le contrat pour la persistance des utilisateurs
// Cette interface appartient au DOMAIN (règles métier)
// Les implémentations seront dans INFRASTRUCTURE
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	GetById(ctx context.Context, id int) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	isEmailTaken(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, user *entities.User) (*entities.User, error)
	DeleteById(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
	Count(ctx context.Context) (int, error)
}

type UserRepositoryFilters struct {
	Email     string
	Name      string
	CreatedAt struct {
		From *string
		To   *string
	}
	Limit  int
	Offset int
}

type UserSearchRepository interface {
	UserRepository
	Search(ctx context.Context, filters UserRepositoryFilters) ([]*entities.User, error)
}
