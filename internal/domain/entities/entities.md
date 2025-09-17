# Domain Entities - Clean Architecture Go

## 📖 Qu'est-ce qu'une Entité ?

Les **Entités** représentent le cœur métier de votre application. Elles encapsulent :
- Les **données essentielles** de votre domaine
- Les **règles métier fondamentales**
- Les **invariants** qui doivent toujours être respectés

Dans la Clean Architecture, les entités sont au **centre du cercle** - elles ne dépendent de rien d'autre et constituent la logique métier la plus stable.

## 🎯 Pourquoi mettre des fonctions métier dans les Entités ?

### ✅ **Avantages :**

1. **Encapsulation** : Les données et les règles qui les gouvernent sont au même endroit
2. **Cohérence** : Impossible de créer un objet dans un état invalide
3. **Réutilisabilité** : Les règles métier sont centralisées et réutilisables
4. **Testabilité** : Facile de tester les règles métier isolément
5. **Expressivité** : Le code reflète le langage métier

### 🔒 **Principe d'encapsulation :**
```go
// ❌ MAUVAIS : Logique métier éparpillée
user.Email = "JOHN@EXAMPLE.COM"  // Pas de validation, pas de normalisation

// ✅ BON : Logique encapsulée dans l'entité
user.UpdateProfile(name, email)  // Validation + normalisation automatique
```

## 🏗️ Structure d'une Entité Go

```go
type User struct {
    // Données de l'entité
    ID       int       `json:"id"`
    Email    string    `json:"email"`
    Name     string    `json:"name"`
    Created  time.Time `json:"created"`
}

// Constructeur avec validation
func NewUser(email, name string) (*User, error) { }

// Méthodes métier publiques (PascalCase)
func (u *User) UpdateProfile(name, email string) error { }
func (u *User) ChangePassword(newPassword string) error { }
func (u *User) IsValid() bool { }

// Méthodes de validation privées (camelCase)
func validateEmail(email string) error { }
func validateName(name string) error { }
```

## 🔤 Conventions de nommage Go

### **Visibilité (Public/Privé) :**
```go
// PUBLIC - exporté vers autres packages
func ValidateEmail(email string) error { }    // PascalCase
type UserRepository interface { }
var GlobalConfig string

// PRIVÉ - visible seulement dans ce package
func validateEmail(email string) error { }    // camelCase  
type userValidator struct { }
var localConfig string
```

**🔑 Règle simple :** `Majuscule = Public`, `minuscule = Privé`

### **Acronymes :**
```go
// ✅ Correct
type HTTPClient struct { }
func ParseJSON() { }
var APIKey string

// ❌ Incorrect  
type HttpClient struct { }
func ParseJson() { }
var ApiKey string
```

## 🛠️ Méthodes avec Receivers

### **Syntaxe :**
```go
func (u *User) UpdateProfile(name string, email string) error
//   ↑     ↑        ↑           ↑                        ↑
//   |     |        |           |                        |
// receiver type   nom      paramètres            retour
```

### **Pointeur vs Valeur :**
```go
// Receiver par pointeur (*User) - MODIFIE l'objet original
func (u *User) UpdateProfile(name string) error {
    u.Name = name  // ✅ Modifie l'original
    return nil
}

// Receiver par valeur (User) - COPIE l'objet  
func (u User) GetDisplayName() string {
    return u.Name  // ✅ Lecture seule, pas de modification
}
```

**🎯 Règle :** Utilisez `*Type` quand vous modifiez, `Type` pour la lecture seule.

## 🔍 Validation avec Regex

### **Compilation de Regex :**
```go
// ❌ Mauvais : Recompile à chaque appel
func validateEmail(email string) error {
    regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !regex.MatchString(email) {
        return errors.New("email invalide")
    }
    return nil
}

// ✅ Bon : Compile une seule fois
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return errors.New("email invalide")
    }
    return nil
}
```

### **Patterns utiles :**
```go
var (
    emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    phoneRegex     = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
    nameRegex      = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s\-'\.]+$`)
    offensiveRegex = regexp.MustCompile(`(?i)(badword1|badword2)`)  // (?i) = insensible à la casse
)
```

## 🧪 Exemple complet d'Entité

```go
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

// Regex compilées une seule fois (performance)
var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    nameRegex  = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s\-'\.]+$`)
)

// Constructeur avec validation
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

// Méthodes métier publiques
func (u *User) UpdateProfile(name, email string) error {
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

func (u *User) IsValid() bool {
    return validateEmail(u.Email) == nil &&
           validateName(u.Name) == nil &&
           validatePassword(u.Password) == nil
}

// Méthodes de validation privées
func validateEmail(email string) error {
    email = strings.TrimSpace(email)
    if email == "" {
        return errors.New("email requis")
    }
    if !emailRegex.MatchString(email) {
        return errors.New("format email invalide")
    }
    return nil
}

func validateName(name string) error {
    name = strings.TrimSpace(name)
    if name == "" {
        return errors.New("nom requis")
    }
    if len(name) < 2 || len(name) > 100 {
        return errors.New("nom doit faire entre 2 et 100 caractères")
    }
    if !nameRegex.MatchString(name) {
        return errors.New("nom contient des caractères invalides")
    }
    return nil
}

func validatePassword(password string) error {
    if len(password) < 6 {
        return errors.New("mot de passe trop court (min 6 caractères)")
    }
    return nil
}
```

## 🎯 Points clés à retenir

1. **Entité = Données + Règles métier**
2. **Majuscule = Public, minuscule = Privé**
3. **Receivers par pointeur pour modifier, par valeur pour lire**
4. **Compilez les regex une seule fois pour la performance**
5. **Validation dans les constructeurs et méthodes de modification**
6. **Invariants toujours respectés = objet toujours dans un état valide**

Les entités forment la **fondation solide** de votre architecture - elles doivent être robustes, bien testées, et refléter fidèlement votre domaine métier ! 🏛️