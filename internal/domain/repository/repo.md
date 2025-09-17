# Repository Pattern - Clean Architecture Go

## 📖 Qu'est-ce qu'un Repository ?

Le **Repository Pattern** est un patron de conception qui encapsule la logique d'accès aux données. Il crée une **couche d'abstraction** entre votre logique métier et votre système de persistance.

Dans la Clean Architecture, les repositories :
- **Interfaces définies dans le Domain** (règles métier)
- **Implémentations dans l'Infrastructure** (détails techniques)
- **Respectent l'inversion de dépendance**

## 🎯 Pourquoi utiliser le Repository Pattern ?

### ✅ **Avantages :**

1. **Abstraction** : La logique métier ne connaît pas les détails de persistance
2. **Testabilité** : Facilite les mocks et tests unitaires
3. **Flexibilité** : Changer de base de données sans impacter le métier
4. **Séparation des responsabilités** : Chaque couche a son rôle
5. **Réutilisabilité** : Interface commune pour différentes implémentations

### 🔄 **Principe d'inversion de dépendance :**
```go
// ❌ MAUVAIS : Use case dépend de l'implémentation
type CreateUserUseCase struct {
    postgresRepo *PostgresUserRepository // Couplage fort !
}

// ✅ BON : Use case dépend de l'interface
type CreateUserUseCase struct {
    userRepo repositories.UserRepository // Abstraction !
}
```

## 🏗️ Structure du Repository

### **Interface (Domain Layer)**
```go
// internal/domain/repositories/user_repository.go
package repositories

import (
    "clean-archi-analytics/internal/domain/entities"
    "context"
)

type UserRepository interface {
    // CRUD de base
    Create(ctx context.Context, user *entities.User) (*entities.User, error)
    GetByID(ctx context.Context, id int) (*entities.User, error)
    Update(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id int) error
    
    // Opérations métier spécifiques
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    
    // Listing et pagination
    List(ctx context.Context, limit, offset int) ([]*entities.User, error)
    Count(ctx context.Context) (int, error)
}
```

### **Implémentation (Infrastructure Layer)**
```go
// internal/infrastructure/database/postgres_user_repository.go
type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
    query := `INSERT INTO users (email, name, password, created, updated) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
    err := r.db.QueryRowContext(ctx, query, user.Email, user.Name, user.Password, 
                               user.Created, user.Updated).Scan(&user.ID)
    return user, err
}
```

## 🧠 Gestion Mémoire - Points Critiques

### **🔍 Comprendre les pointeurs dans les repositories**

#### **1. Signatures avec pointeurs :**
```go
// Pourquoi *entities.User et pas entities.User ?

// ✅ Avec pointeur - Efficace
func (r *Repo) Update(user *entities.User) error
// → Passe l'adresse mémoire, pas de copie

// ❌ Sans pointeur - Inefficace  
func (r *Repo) Update(user entities.User) error
// → Copie tout l'objet à chaque appel
```

#### **2. Slice de pointeurs `[]*entities.User` :**
```go
func (r *Repo) List() ([]*entities.User, error)
// → Retourne un slice contenant des ADRESSES vers des User
// → Plus efficace qu'un slice de copies []entities.User
```

### **⚠️ Piège : Références partagées**

```go
// ❌ DANGEREUX - Retourne les pointeurs originaux
func (r *MockRepo) List() []*entities.User {
    var users []*entities.User
    for _, user := range r.storage {
        users = append(users, user) // MÊME pointeur !
    }
    return users // ⚠️ Modifications externes possibles !
}

// ✅ SÉCURISÉ - Retourne des copies
func (r *MockRepo) List() []*entities.User {
    var users []*entities.User
    for _, user := range r.storage {
        userCopy := *user  // Copie la VALEUR
        users = append(users, &userCopy) // NOUVEAU pointeur
    }
    return users // ✅ Modifications externes sans impact
}
```

### **🔧 Update : Modifier en place vs Remplacer**

```go
// Option A: MODIFICATION EN PLACE (efficace)
func (r *Repo) Update(user *entities.User) error {
    existing := r.storage[user.ID]
    existing.Name = user.Name     // Même adresse mémoire
    existing.Email = user.Email   // ✅ Économique
    existing.Updated = time.Now()
    return nil
}

// Option B: REMPLACEMENT (simple mais coûteux)
func (r *Repo) Update(user *entities.User) error {
    userCopy := *user                    // Nouvelle copie
    r.storage[user.ID] = &userCopy      // ❌ Nouvelle allocation
    return nil
}
```

**💡 Règle :** Préférez la modification en place pour la performance, sauf si l'immutabilité est requise.

## 🛡️ Utilisation du Context

### **Pourquoi `context.Context` partout ?**

```go
func (r *UserRepository) GetByID(ctx context.Context, id int) (*entities.User, error)
```

**Avantages du context :**
1. **Timeouts** : Annulation des requêtes longues
2. **Cancellation** : Arrêt propre des opérations
3. **Tracing** : Suivi des requêtes dans les logs
4. **Valeurs partagées** : Données de session, utilisateur connecté

```go
// Exemple d'utilisation avec timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := userRepo.GetByID(ctx, 123)
if err != nil {
    // Peut être dû au timeout ou à l'absence de l'utilisateur
}
```

## 🧪 Testing avec Mocks

### **Mock Repository pour les tests :**

```go
// internal/domain/repositories/mocks/user_repository_mock.go
type MockUserRepository struct {
    users  map[int]*entities.User
    emails map[string]*entities.User // Index pour recherche rapide
    nextID int
    mutex  sync.RWMutex // Protection concurrence
}

func NewMockUserRepository() repositories.UserRepository {
    return &MockUserRepository{
        users:  make(map[int]*entities.User),
        emails: make(map[string]*entities.User),
        nextID: 1,
    }
}
```

### **Points clés du mock :**

```go
func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    // 1. Validation métier (email unique)
    if _, exists := m.emails[user.Email]; exists {
        return nil, errors.New("email déjà utilisé")
    }
    
    // 2. Génération ID auto-increment
    user.ID = m.nextID
    m.nextID++
    
    // 3. Copie pour éviter modifications externes
    userCopy := *user
    
    // 4. Double indexation (ID + email)
    m.users[user.ID] = &userCopy
    m.emails[user.Email] = &userCopy
    
    return &userCopy, nil
}
```

## 🎭 Exemple complet d'utilisation

### **1. Dans le Use Case :**
```go
type CreateUserUseCase struct {
    userRepo repositories.UserRepository
    hasher   PasswordHasher
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, email, name, password string) (*entities.User, error) {
    // 1. Vérifier si l'email existe
    exists, err := uc.userRepo.ExistsByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email déjà utilisé")
    }
    
    // 2. Créer l'entité avec validation
    user, err := entities.NewUser(email, name, password)
    if err != nil {
        return nil, err
    }
    
    // 3. Hasher le mot de passe
    user.Password = uc.hasher.Hash(user.Password)
    
    // 4. Sauvegarder via le repository
    return uc.userRepo.Create(ctx, user)
}
```

### **2. Dans les tests :**
```go
func TestCreateUser(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository()
    mockHasher := &MockPasswordHasher{}
    useCase := NewCreateUserUseCase(mockRepo, mockHasher)
    
    // Act
    user, err := useCase.Execute(context.Background(), "test@example.com", "John", "password123")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
    assert.NotZero(t, user.ID)
}
```

## 📊 Patterns avancés

### **Repository avec filtres :**
```go
type UserFilters struct {
    Email     string
    Name      string
    CreatedAt struct {
        From *time.Time
        To   *time.Time
    }
    Limit  int
    Offset int
}

type UserSearchRepository interface {
    UserRepository
    Search(ctx context.Context, filters UserFilters) ([]*entities.User, error)
}
```

### **Repository avec transactions :**
```go
type TransactionalUserRepository interface {
    UserRepository
    WithTx(tx *sql.Tx) UserRepository
}
```

## 🎯 Bonnes pratiques

### **✅ À faire :**
1. **Interfaces dans le domain**, implémentations dans l'infrastructure
2. **Utiliser context.Context** pour timeouts et cancellation
3. **Retourner des copies** dans List() et GetByID() pour la sécurité
4. **Validation au niveau métier** (email unique, etc.)
5. **Gestion d'erreurs explicite** (pas de nil silencieux)
6. **Mocks pour les tests** avec état en mémoire

### **❌ À éviter :**
1. Logique métier dans le repository (seule la persistance)
2. Retourner `nil, nil` (ambigu)
3. Ignorer les erreurs de context (timeout)
4. Partager des pointeurs sans protection
5. Repository trop spécialisé (une méthode par cas d'usage)

## 🔗 Intégration avec les autres couches

```
Use Cases (Domain) 
    ↓ utilise interface
Repository Interface (Domain)
    ↑ implémente
Repository Implementation (Infrastructure)
    ↓ utilise
Database (Infrastructure)
```

Le repository fait le **pont** entre votre logique métier pure et votre système de persistance, en respectant les principes de la Clean Architecture ! 🏛️

## 📝 Mémo Gestion Mémoire

| Opération | Allocation | Sécurité | Performance |
|-----------|------------|----------|-------------|
| `List()` avec copies | ➕ Nouvelle | ✅ Haute | ➖ Moyenne |
| `List()` avec refs | ➖ Aucune | ❌ Faible | ✅ Haute |
| `Update()` en place | ➖ Aucune | ✅ Haute | ✅ Haute |
| `Update()` avec copie | ➕ Nouvelle | ✅ Haute | ➖ Moyenne |

**Recommandation :** Copies pour la lecture, modification en place pour l'écriture.