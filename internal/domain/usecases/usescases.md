# Use Cases - Clean Architecture Go

## 📖 Qu'est-ce qu'un Use Case ?

Les **Use Cases** représentent les **scénarios métier** de votre application. Ils orchestrent les entités et repositories pour implémenter des fonctionnalités complètes selon les règles business.

Dans la Clean Architecture, les Use Cases :
- **Contiennent la logique applicative** (comment faire)
- **Orchestrent les entités** (règles métier fondamentales)
- **Utilisent les repositories** (persistance)
- **Coordonnent les services externes** (email, logs, etc.)
- **Transforment les données** (DTOs input/output)

## 🎯 Use Cases = Business Facades

Les Use Cases utilisent le **pattern Facade** mais l'enrichissent avec de la logique métier :

### **🔄 Facade classique vs Use Case**

```go
// ❌ Facade simple - Juste une agrégation
type DatabaseFacade struct {
    userRepo UserRepository
    orderRepo OrderRepository
}

func (f *DatabaseFacade) GetUserWithOrders(userID int) (*UserWithOrders, error) {
    user := f.userRepo.GetByID(userID)        // Appel direct
    orders := f.orderRepo.GetByUserID(userID) // Appel direct
    return &UserWithOrders{user, orders}, nil // Combinaison basique
}

// ✅ Use Case - Facade + Logique métier + Orchestration
type CreateUserUseCase struct {
    userRepo     repositories.UserRepository
    passwordHash PasswordHasher
    emailSender  EmailSender
    logger       Logger
}

func (uc *CreateUserUseCase) Execute(req CreateUserRequest) (*CreateUserResponse, error) {
    // 1. RÈGLES MÉTIER
    if exists := uc.userRepo.ExistsByEmail(req.Email); exists {
        return nil, errors.New("email déjà utilisé") // BUSINESS RULE
    }
    
    // 2. VALIDATION MÉTIER
    user, err := entities.NewUser(req.Email, req.Name, req.Password)
    if err != nil {
        return nil, err // Validation des entités
    }
    
    // 3. ORCHESTRATION AVEC LOGIQUE
    hashedPassword := uc.passwordHash.Hash(user.Password)
    user.Password = hashedPassword
    
    // 4. PERSISTANCE
    createdUser := uc.userRepo.Create(ctx, user)
    
    // 5. EFFETS DE BORD MÉTIER
    go uc.emailSender.SendWelcomeEmail(user.Email, user.Name)
    
    // 6. TRANSFORMATION OUTPUT
    return toCreateUserResponse(createdUser), nil
}
```

## 🏗️ Anatomie d'un Use Case

### **Structure standard :**

```go
type CreateUserUseCase struct {
    // Dependencies (Inversion de dépendance)
    userRepo     repositories.UserRepository // Interface, pas implémentation
    passwordHash PasswordHasher               // Service externe
    emailSender  EmailSender                  // Service externe  
    logger       Logger                       // Service externe
}

// Constructeur avec injection de dépendances
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

// DTOs spécifiques au Use Case
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=6"`
}

type CreateUserResponse struct {
    ID      int       `json:"id"`
    Email   string    `json:"email"`
    Name    string    `json:"name"`
    Created time.Time `json:"created"`
    // Pas de Password dans la réponse !
}

// Méthode principale qui implémente le scénario métier
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    // Logique métier complète ici
}
```

## 🔀 Patterns utilisés dans les Use Cases

### **1. Command Pattern**
Chaque Use Case encapsule une **commande métier** complète :

```go
// Une commande = Une intention métier
type CreateUserUseCase struct { /* ... */ }  // "Créer un utilisateur"
type UpdateUserUseCase struct { /* ... */ }  // "Modifier un utilisateur"  
type DeleteUserUseCase struct { /* ... */ }  // "Supprimer un utilisateur"
type SendPasswordResetUseCase struct { /* ... */ } // "Envoyer reset password"

// Chaque commande a sa méthode Execute()
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
```

### **2. Facade Pattern (enrichi)**
Interface simple pour orchestrer plusieurs services complexes :

```go
// Client voit une interface simple
userService.CreateUser(email, name, password)

// Mais en interne, orchestration complexe :
// → Validation entité
// → Vérification email unique  
// → Hash password
// → Sauvegarde DB
// → Envoi email welcome
// → Logging
// → Transformation réponse
```

### **3. DTO Pattern**
Séparation claire entre données externes et entités internes :

```go
// ✅ Input DTO - Données du client
type CreateUserRequest struct {
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"password"`
}

// ✅ Entity - Logique métier interne
type User struct {
    ID       int       
    Email    string    
    Name     string    
    Password string    // Hashé
    Created  time.Time 
    Updated  time.Time 
}

// ✅ Output DTO - Données pour le client (sans password)
type CreateUserResponse struct {
    ID      int       `json:"id"`
    Email   string    `json:"email"`
    Name    string    `json:"name"`
    Created time.Time `json:"created"`
}
```

### **4. Dependency Inversion**
Use Cases dépendent d'**interfaces**, jamais d'implémentations :

```go
// ✅ Dépend de l'interface (domain)
type CreateUserUseCase struct {
    userRepo repositories.UserRepository // Interface
}

// ❌ Ne dépend PAS de l'implémentation (infrastructure)
type CreateUserUseCase struct {
    userRepo *PostgresUserRepository // Implémentation concrète
}
```

## 🎭 Responsabilités des Use Cases

### **✅ Ce qu'ils FONT :**

1. **Orchestration** : Coordonner entités, repositories et services
2. **Validation applicative** : Règles business spécifiques au cas d'usage
3. **Transformation des données** : DTOs ↔ Entités
4. **Gestion des erreurs métier** : Erreurs compréhensibles pour le client
5. **Logging et monitoring** : Observabilité des opérations métier
6. **Coordination des effets de bord** : Emails, notifications, etc.

```go
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    // 1. Validation applicative
    exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
    if exists {
        return nil, errors.New("email déjà utilisé") // Erreur métier
    }
    
    // 2. Orchestration entités
    user, err := entities.NewUser(req.Email, req.Name, req.Password)
    
    // 3. Coordination services
    hashedPassword, err := uc.passwordHash.Hash(user.Password)
    user.Password = hashedPassword
    
    // 4. Persistance
    createdUser, err := uc.userRepo.Create(ctx, user)
    
    // 5. Effets de bord (asynchrones)
    go uc.emailSender.SendWelcomeEmail(ctx, createdUser.Email, createdUser.Name)
    
    // 6. Logging
    uc.logger.Info("User created", map[string]interface{}{
        "user_id": createdUser.ID,
        "email":   createdUser.Email,
    })
    
    // 7. Transformation output
    return &CreateUserResponse{
        ID:      createdUser.ID,
        Email:   createdUser.Email,
        Name:    createdUser.Name,
        Created: createdUser.Created,
    }, nil
}
```

### **❌ Ce qu'ils ne font PAS :**

1. **Logique métier fondamentale** → Dans les entités
2. **Détails de persistance** → Dans les repositories
3. **Logique de présentation** → Dans les handlers/controllers
4. **Configuration technique** → Dans l'infrastructure

## 🧪 Testing des Use Cases

### **Facilité de test grâce aux mocks :**

```go
func TestCreateUserUseCase_Success(t *testing.T) {
    // Arrange - Mocks de toutes les dépendances
    mockRepo := mocks.NewMockUserRepository()
    mockHasher := &MockPasswordHasher{}
    mockEmailSender := &MockEmailSender{}
    mockLogger := &MockLogger{}
    
    useCase := NewCreateUserUseCase(mockRepo, mockHasher, mockEmailSender, mockLogger)
    
    // Définir le comportement attendu des mocks
    mockHasher.On("Hash", "password123").Return("hashed_password", nil)
    mockEmailSender.On("SendWelcomeEmail", mock.Anything, "john@test.com", "John").Return(nil)
    
    // Act
    req := CreateUserRequest{
        Email:    "john@test.com",
        Name:     "John Doe", 
        Password: "password123",
    }
    response, err := useCase.Execute(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "john@test.com", response.Email)
    assert.NotZero(t, response.ID)
    
    // Vérifier que les mocks ont été appelés correctement
    mockHasher.AssertExpectations(t)
}
```

### **Tests des cas d'erreur :**

```go
func TestCreateUserUseCase_EmailAlreadyExists(t *testing.T) {
    mockRepo := mocks.NewMockUserRepository()
    // ... setup mocks
    
    // Pré-créer un utilisateur avec le même email
    existingUser, _ := entities.NewUser("john@test.com", "Existing", "pass")
    mockRepo.Create(context.Background(), existingUser)
    
    useCase := NewCreateUserUseCase(mockRepo, mockHasher, mockEmailSender, mockLogger)
    
    req := CreateUserRequest{Email: "john@test.com", Name: "John", Password: "pass"}
    
    // Act
    response, err := useCase.Execute(context.Background(), req)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, response)
    assert.Contains(t, err.Error(), "déjà utilisé")
}
```

## 🚀 Use Cases avancés

### **1. Use Case avec pagination :**

```go
type ListUsersUseCase struct {
    userRepo repositories.UserRepository
    logger   Logger
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
    if req.Page == 0 { req.Page = 1 }
    if req.PageSize == 0 { req.PageSize = 10 }
    
    // Calcul offset
    offset := (req.Page - 1) * req.PageSize
    
    // Récupération données + comptage
    users, err := uc.userRepo.List(ctx, req.PageSize, offset)
    total, err := uc.userRepo.Count(ctx)
    
    // Transformation + calculs pagination
    totalPages := (total + req.PageSize - 1) / req.PageSize
    
    return &ListUsersResponse{
        Users:      convertToUserResponses(users),
        Total:      total,
        Page:       req.Page,
        PageSize:   req.PageSize,
        TotalPages: totalPages,
    }, nil
}
```

### **2. Use Case avec transaction :**

```go
type TransferMoneyUseCase struct {
    accountRepo repositories.AccountRepository
    txManager   TransactionManager
    logger      Logger
}

func (uc *TransferMoneyUseCase) Execute(ctx context.Context, fromID, toID int, amount float64) error {
    return uc.txManager.WithTransaction(ctx, func(ctx context.Context) error {
        // 1. Vérifier solde suffisant
        fromAccount, err := uc.accountRepo.GetByID(ctx, fromID)
        if fromAccount.Balance < amount {
            return errors.New("solde insuffisant")
        }
        
        // 2. Effectuer le transfert (logique métier dans les entités)
        fromAccount.Withdraw(amount)
        
        toAccount, err := uc.accountRepo.GetByID(ctx, toID)
        toAccount.Deposit(amount)
        
        // 3. Sauvegarder les deux comptes (dans la même transaction)
        err = uc.accountRepo.Update(ctx, fromAccount)
        err = uc.accountRepo.Update(ctx, toAccount)
        
        return nil // Commit automatique si pas d'erreur
    })
}
```

## 📊 Use Cases vs autres couches

| Couche | Responsabilité | Exemple |
|--------|----------------|---------|
| **Entities** | Règles métier fondamentales | `user.UpdateProfile()` |
| **Use Cases** | Scénarios applicatifs | `CreateUserUseCase.Execute()` |
| **Repositories** | Persistance des données | `userRepo.Create()` |
| **Handlers** | Interface HTTP/API | `POST /users` |
| **Services** | Utilitaires techniques | `passwordHasher.Hash()` |

## 🎯 Bonnes pratiques

### **✅ À faire :**

1. **Un Use Case = Un scénario métier** (pas trop granulaire)
2. **DTOs spécifiques** pour chaque Use Case (input/output)
3. **Gestion d'erreurs métier** (pas d'erreurs techniques qui remontent)
4. **Injection de dépendances** via constructeur
5. **Logging approprié** pour l'observabilité
6. **Tests unitaires** avec mocks
7. **Context propagation** pour timeouts/cancellation

### **❌ À éviter :**

1. **Logique métier dans les Use Cases** → Dans les entités
2. **Use Cases trop fins** (un par méthode CRUD)
3. **Dépendances concrètes** → Toujours des interfaces
4. **Retourner des entités directement** → Utiliser des DTOs
5. **Ignorer les erreurs** ou les propager sans transformation
6. **Use Cases stateful** → Toujours stateless

## 🔗 Intégration avec les autres couches

```
Controllers (Infrastructure)
    ↓ appelle
Use Cases (Domain)
    ↓ utilise
Entities (Domain) + Repositories (Domain interfaces)
    ↓ implémentées par
Repository Implementations (Infrastructure)
```

## 💡 En résumé

Les **Use Cases** sont le **chef d'orchestre** de votre application :

- Ils **orchestrent** les entités et services
- Ils **implémentent** les scénarios métier complets
- Ils **transforment** les données entre couches
- Ils **gèrent** les erreurs et effets de bord
- Ils **facilitent** les tests et la maintenance

C'est la couche qui donne **du sens** à votre application en combinant tous les éléments techniques pour créer de la **valeur métier** ! 🎯