package entities

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Password string    `json:"password,omitempty"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

func NewUser(email, name, password string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Name:     strings.TrimSpace(name),
		Password: password,
		Created:  now,
		Updated:  now,
	}, nil
}

/*
Comprendre les fonctions avec déclarations et receiver :
func : mot-clé pour déclarer une fonction
(u *User) : receiver - cette fonction appartient au type User
u : nom de la variable (comme self en Python)
*User : pointeur vers User (la fonction peut modifier l'objet)
UpdateUserProfile : nom de la fonction
(name string, email string) : paramètres
error : type de retour
*/
func (u *User) UpdateUserProfile(name string, email string) error {
	if err := validateName(name); err != nil {
		return err
	}

	if err := validateEmail(email); err != nil {
		return err
	}

	u.Name = strings.TrimSpace(name)
	u.Email = strings.ToLower(strings.TrimSpace(email))
	u.Updated = time.Now()

	return nil
}

func (u *User) ChangePassword(newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	u.Password = newPassword // À hasher dans le use case
	u.Updated = time.Now()

	return nil
}

func (u *User) isValidUser() bool {
	return validateEmail(u.Email) == nil &&
		validateName(u.Name) == nil &&
		validatePassword(u.Password) == nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email can't be empty")
	}

	if !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}

	if len(email) > 255 {
		return errors.New("email too long")
	}

	return nil
}

var (
	offensiveNameRegex = regexp.MustCompile(`(?i)(fuck|shit|damn|idiot|stupid|hitler|cunt)`)
	validNameRegex     = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s\-'.]+$`)
)

func validateName(name string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return errors.New("empty name")
	}

	if len(trimmedName) < 2 {
		return errors.New("brother nobody has such a short name")
	}

	if len(trimmedName) > 100 {
		return errors.New("brother nobody has such a long name")
	}

	if offensiveNameRegex.MatchString(trimmedName) {
		return errors.New("offensive name non approprié")
	}

	if !validNameRegex.MatchString(trimmedName) {
		return errors.New("nom contient des caractères invalides")
	}

	return nil
}
func validatePassword(password string) error {
	if password == "" {
		return errors.New("mot de passe ne peut pas être vide")
	}

	if len(password) < 6 {
		return errors.New("mot de passe trop court (min 6 caractères)")
	}

	if len(password) > 128 {
		return errors.New("mot de passe trop long (max 128 caractères)")
	}

	return nil
}
