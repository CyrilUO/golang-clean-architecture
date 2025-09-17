// internal/domain/usecases/user_usecases.go
package usecases

import (
	"clean-archi-analytics/internal/domain/entities"
	"clean-archi-analytics/internal/domain/repositories"
	"context"
	"errors"
	"time"
)

// =============================================================================
// INTERFACES POUR LES SERVICES EXTERNES (Dependency Inversion)
// =============================================================================

// PasswordHasher interface pour hasher les mots de passe
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// EmailSender interface pour envoyer des emails
type EmailSender interface {
	SendWelcomeEmail(ctx context.Context, email, name string) error
}

// Logger interface pour les logs
type Logger interface {
	Info(message string, fields map[string]interface{})
	Error(message string, err error, fields map[string]interface{})
}

// =============================================================================
// CREATE USER USE CASE
// =============================================================================

type CreateUserUseCase struct {
	userRepo     repositories.UserRepository
	passwordHash PasswordHasher
	emailSender  EmailSender
	logger       Logger
}

func NewCreateUserUseCase(
	userRepo repositories.UserRepository,
	passwordHash PasswordHasher,
	emailSender EmailSender,
	logger Logger,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:     userRepo,
		passwordHash: passwordHash,
		emailSender:  emailSender,
		logger:       logger,
	}
}

// CreateUserRequest DTO pour l'input
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

// CreateUserResponse DTO pour l'output
type CreateUserResponse struct {
	ID      int       `json:"id"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	uc.logger.Info("Creating new user", map[string]interface{}{
		"email": req.Email,
		"name":  req.Name,
	})

	// 1. Vérifier que l'email n'existe pas déjà
	exists, err := uc.userRepo.isEmailTaken(ctx, req.Email)
	if err != nil {
		uc.logger.Error("Failed to check email existence", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, errors.New("erreur lors de la vérification de l'email")
	}

	if exists {
		return nil, errors.New("un utilisateur avec cet email existe déjà")
	}

	// 2. Créer l'entité User avec validation métier
	user, err := entities.NewUser(req.Email, req.Name, req.Password)
	if err != nil {
		uc.logger.Error("Failed to create user entity", err, map[string]interface{}{
			"email": req.Email,
			"name":  req.Name,
		})
		return nil, err
	}

	// 3. Hasher le mot de passe
	hashedPassword, err := uc.passwordHash.Hash(user.Password)
	if err != nil {
		uc.logger.Error("Failed to hash password", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, errors.New("erreur lors du traitement du mot de passe")
	}
	user.Password = hashedPassword

	// 4. Sauvegarder en base
	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		uc.logger.Error("Failed to save user", err, map[string]interface{}{
			"email": req.Email,
			"name":  req.Name,
		})
		return nil, errors.New("erreur lors de la création de l'utilisateur")
	}

	// 5. Envoyer email de bienvenue (asynchrone, ne doit pas faire échouer la création)
	go func() {
		if err := uc.emailSender.SendWelcomeEmail(context.Background(), createdUser.Email, createdUser.Name); err != nil {
			uc.logger.Error("Failed to send welcome email", err, map[string]interface{}{
				"user_id": createdUser.ID,
				"email":   createdUser.Email,
			})
		}
	}()

	uc.logger.Info("User created successfully", map[string]interface{}{
		"user_id": createdUser.ID,
		"email":   createdUser.Email,
	})

	// 6. Retourner la réponse (sans le mot de passe)
	return &CreateUserResponse{
		ID:      createdUser.ID,
		Email:   createdUser.Email,
		Name:    createdUser.Name,
		Created: createdUser.Created,
	}, nil
}

// =============================================================================
// GET USER USE CASE
// =============================================================================

type GetUserUseCase struct {
	userRepo repositories.UserRepository
	logger   Logger
}

func NewGetUserUseCase(userRepo repositories.UserRepository, logger Logger) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

type GetUserResponse struct {
	ID      int       `json:"id"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func (uc *GetUserUseCase) ExecuteByID(ctx context.Context, id int) (*GetUserResponse, error) {
	user, err := uc.userRepo.GetById(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user by ID", err, map[string]interface{}{
			"user_id": id,
		})
		return nil, errors.New("utilisateur non trouvé")
	}

	return &GetUserResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Created: user.Created,
		Updated: user.Updated,
	}, nil
}

func (uc *GetUserUseCase) ExecuteByEmail(ctx context.Context, email string) (*GetUserResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		uc.logger.Error("Failed to get user by email", err, map[string]interface{}{
			"email": email,
		})
		return nil, errors.New("utilisateur non trouvé")
	}

	return &GetUserResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Created: user.Created,
		Updated: user.Updated,
	}, nil
}

// =============================================================================
// UPDATE USER USE CASE
// =============================================================================

type UpdateUserUseCase struct {
	userRepo repositories.UserRepository
	logger   Logger
}

func NewUpdateUserUseCase(userRepo repositories.UserRepository, logger Logger) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

type UpdateUserRequest struct {
	ID    int    `json:"id" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateUserResponse struct {
	ID      int       `json:"id"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	Updated time.Time `json:"updated"`
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
	uc.logger.Info("Updating user", map[string]interface{}{
		"user_id": req.ID,
		"email":   req.Email,
		"name":    req.Name,
	})

	// 1. Récupérer l'utilisateur existant
	user, err := uc.userRepo.GetById(ctx, req.ID)
	if err != nil {
		uc.logger.Error("Failed to get user for update", err, map[string]interface{}{
			"user_id": req.ID,
		})
		return nil, errors.New("utilisateur non trouvé")
	}

	// 2. Si l'email change, vérifier qu'il n'est pas pris
	if user.Email != req.Email {
		exists, err := uc.userRepo.isEmailTaken(ctx, req.Email)
		if err != nil {
			uc.logger.Error("Failed to check email existence for update", err, map[string]interface{}{
				"email": req.Email,
			})
			return nil, errors.New("erreur lors de la vérification de l'email")
		}

		if exists {
			return nil, errors.New("cet email est déjà utilisé")
		}
	}

	// 3. Utiliser la méthode métier de l'entité pour la mise à jour
	if err := user.UpdateUserProfile(req.Name, req.Email); err != nil {
		uc.logger.Error("Failed to update user profile", err, map[string]interface{}{
			"user_id": req.ID,
		})
		return nil, err
	}

	// 4. Sauvegarder les modifications
	if _, err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.Error("Failed to save user update", err, map[string]interface{}{
			"user_id": req.ID,
		})
		return nil, errors.New("erreur lors de la mise à jour")
	}

	uc.logger.Info("User updated successfully", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})

	return &UpdateUserResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Updated: user.Updated,
	}, nil
}

// =============================================================================
// DELETE USER USE CASE
// =============================================================================

type DeleteUserUseCase struct {
	userRepo repositories.UserRepository
	logger   Logger
}

func NewDeleteUserUseCase(userRepo repositories.UserRepository, logger Logger) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uc *DeleteUserUseCase) Execute(ctx context.Context, id int) error {
	uc.logger.Info("Deleting user", map[string]interface{}{
		"user_id": id,
	})

	// 1. Vérifier que l'utilisateur existe
	_, err := uc.userRepo.GetById(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get user for deletion", err, map[string]interface{}{
			"user_id": id,
		})
		return errors.New("utilisateur non trouvé")
	}

	// 2. Supprimer l'utilisateur
	if err := uc.userRepo.DeleteById(ctx, id); err != nil {
		uc.logger.Error("Failed to delete user", err, map[string]interface{}{
			"user_id": id,
		})
		return errors.New("erreur lors de la suppression")
	}

	uc.logger.Info("User deleted successfully", map[string]interface{}{
		"user_id": id,
	})

	return nil
}

// =============================================================================
// LIST USERS USE CASE (avec pagination)
// =============================================================================

type ListUsersUseCase struct {
	userRepo repositories.UserRepository
	logger   Logger
}

func NewListUsersUseCase(userRepo repositories.UserRepository, logger Logger) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

type ListUsersRequest struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

type ListUsersResponse struct {
	Users      []*GetUserResponse `json:"users"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// Valeurs par défaut
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Calculer offset
	offset := (req.Page - 1) * req.PageSize

	// Récupérer les utilisateurs
	users, err := uc.userRepo.List(ctx, req.PageSize, offset)
	if err != nil {
		uc.logger.Error("Failed to list users", err, map[string]interface{}{
			"page":      req.Page,
			"page_size": req.PageSize,
		})
		return nil, errors.New("erreur lors de la récupération des utilisateurs")
	}

	// Compter le total
	total, err := uc.userRepo.Count(ctx)
	if err != nil {
		uc.logger.Error("Failed to count users", err, nil)
		return nil, errors.New("erreur lors du comptage des utilisateurs")
	}

	// Convertir en DTO
	userResponses := make([]*GetUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &GetUserResponse{
			ID:      user.ID,
			Email:   user.Email,
			Name:    user.Name,
			Created: user.Created,
			Updated: user.Updated,
		}
	}

	// Calculer le nombre de pages
	totalPages := (total + req.PageSize - 1) / req.PageSize

	return &ListUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
