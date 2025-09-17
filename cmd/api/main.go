package main

import (
	"fmt"
)

type User struct {
	ID   int
	Name string
}

// =============================================================================
// 1. DIFFÉRENCE ENTRE POINTEUR ET COPIE
// =============================================================================

func demonstratePointerVsCopy() {
	fmt.Println("=== 1. POINTEUR VS COPIE ===")

	// Créer un utilisateur
	user := User{ID: 1, Name: "Alice"}
	fmt.Printf("user original: %+v (adresse: %p)\n", user, &user)

	// COPIE : Passer par valeur
	modifyByCopy(user)
	fmt.Printf("après modifyByCopy: %+v (pas changé!)\n", user)

	// POINTEUR : Passer par référence
	modifyByPointer(&user)
	fmt.Printf("après modifyByPointer: %+v (changé!)\n", user)
}

func modifyByCopy(u User) {
	fmt.Printf("  dans modifyByCopy: %p (adresse différente!)\n", &u)
	u.Name = "Bob" // Modifie la COPIE, pas l'original
}

func modifyByPointer(u *User) {
	fmt.Printf("  dans modifyByPointer: %p (même adresse!)\n", u)
	u.Name = "Charlie" // Modifie l'ORIGINAL
}

// =============================================================================
// 2. CAS DU REPOSITORY LIST() - SLICE DE POINTEURS
// =============================================================================

func demonstrateList() {
	fmt.Println("\n=== 2. LIST() - SLICE DE POINTEURS ===")

	// Simuler des données en "base"
	storage := map[int]*User{
		1: {ID: 1, Name: "Alice"},
		2: {ID: 2, Name: "Bob"},
	}

	// Version DANGEREUSE : retourne les pointeurs originaux
	dangerousList := func() []*User {
		var users []*User
		for _, user := range storage {
			users = append(users, user) // ⚠️ MÊME POINTEUR !
		}
		return users
	}

	// Version SÉCURISÉE : retourne des copies
	safeList := func() []*User {
		var users []*User
		for _, user := range storage {
			userCopy := *user                // Copie la valeur
			users = append(users, &userCopy) // Nouveau pointeur vers la copie
		}
		return users
	}

	// Test version dangereuse
	dangerousUsers := dangerousList()
	fmt.Printf("Version dangereuse - user[0]: %p\n", dangerousUsers[0])
	fmt.Printf("Storage user[1]: %p\n", storage[1])
	fmt.Printf("Même adresse? %t\n", dangerousUsers[0] == storage[1])

	// Si on modifie via dangerousUsers, on modifie le storage !
	dangerousUsers[0].Name = "MODIFIED!"
	fmt.Printf("Storage après modification: %+v\n", storage[1])

	// Reset
	storage[1].Name = "Bob"

	// Test version sécurisée
	safeUsers := safeList()
	fmt.Printf("\nVersion sécurisée - user[0]: %p\n", safeUsers[0])
	fmt.Printf("Storage user[1]: %p\n", storage[1])
	fmt.Printf("Même adresse? %t\n", safeUsers[0] == storage[1])

	// Modification n'affecte pas le storage
	safeUsers[0].Name = "MODIFIED COPY!"
	fmt.Printf("Storage après modification: %+v (inchangé!)\n", storage[1])
}

// =============================================================================
// 3. CAS DE UPDATE() - MODIFICATION D'OBJET EXISTANT
// =============================================================================

func demonstrateUpdate() {
	fmt.Println("\n=== 3. UPDATE() - MODIFICATION OBJET EXISTANT ===")

	// Simuler le storage
	storage := map[int]*User{
		1: {ID: 1, Name: "Alice"},
	}

	fmt.Printf("Avant update - storage[1]: %+v (adresse: %p)\n",
		storage[1], storage[1])

	// Cas 1: Update qui MODIFIE l'objet existant (économique)
	updateInPlace := func(user *User) {
		existingUser := storage[user.ID]
		// Modifier les champs un par un
		existingUser.Name = user.Name
		// L'objet reste à la même adresse mémoire
	}

	// Cas 2: Update qui REMPLACE l'objet (moins économique)
	updateWithReplace := func(user *User) {
		// Créer une nouvelle copie
		newUser := *user
		storage[user.ID] = &newUser // ⚠️ Nouvelle allocation !
	}

	// Test update in-place
	modifiedUser := User{ID: 1, Name: "Alice Updated"}
	oldAddr := storage[1]
	updateInPlace(&modifiedUser)

	fmt.Printf("Après updateInPlace - storage[1]: %+v (adresse: %p)\n",
		storage[1], storage[1])
	fmt.Printf("Même adresse? %t\n", oldAddr == storage[1])

	// Test update with replace
	modifiedUser2 := User{ID: 1, Name: "Alice Replaced"}
	oldAddr2 := storage[1]
	updateWithReplace(&modifiedUser2)

	fmt.Printf("Après updateWithReplace - storage[1]: %+v (adresse: %p)\n",
		storage[1], storage[1])
	fmt.Printf("Même adresse? %t\n", oldAddr2 == storage[1])
}

// =============================================================================
// 4. ALLOCATION MÉMOIRE - STACK VS HEAP
// =============================================================================

func demonstrateStackVsHeap() {
	fmt.Println("\n=== 4. STACK VS HEAP ===")

	// Variables locales (généralement sur la stack)
	localUser := User{ID: 1, Name: "Local"}
	fmt.Printf("Local user (stack probablement): %p\n", &localUser)

	// Allocation explicite sur le heap avec new()
	heapUser := new(User)
	*heapUser = User{ID: 2, Name: "Heap"}
	fmt.Printf("Heap user (heap): %p\n", heapUser)

	// Allocation avec make pour un slice
	users := make([]*User, 0, 10)
	fmt.Printf("Slice (heap): %p\n", users)

	// Quand une variable locale "s'échappe", Go la met automatiquement sur le heap
	escapeToHeap := func() *User {
		localUser := User{ID: 3, Name: "Escaped"}
		return &localUser // ⚠️ Cette variable va sur le heap automatiquement !
	}

	escapedUser := escapeToHeap()
	fmt.Printf("Escaped user (heap automatiquement): %p\n", escapedUser)
}

// =============================================================================
// 5. EXEMPLES PRATIQUES REPOSITORY
// =============================================================================

// Mock Repository sécurisé
type MockUserRepository struct {
	users map[int]*User
}

// List() - Version sécurisée qui retourne des copies
func (r *MockUserRepository) List() []*User {
	var result []*User
	for _, user := range r.users {
		// Créer une COPIE pour éviter les modifications accidentelles
		userCopy := *user                  // Copie la valeur
		result = append(result, &userCopy) // Nouveau pointeur vers la copie
	}
	return result
}

// Update() - Version qui modifie en place (efficace)
func (r *MockUserRepository) Update(user *User) error {
	existing, exists := r.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Option 1: Modifier en place (même allocation mémoire)
	existing.Name = user.Name
	// existing reste à la même adresse

	// Option 2: Remplacer complètement (nouvelle allocation)
	// userCopy := *user
	// r.users[user.ID] = &userCopy

	return nil
}

// GetByID() - Retourne une copie pour sécurité
func (r *MockUserRepository) GetByID(id int) (*User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Retourner une COPIE pour éviter les modifications externes
	userCopy := *user
	return &userCopy, nil
}

func demonstrateRepositoryMemory() {
	fmt.Println("\n=== 5. REPOSITORY MEMORY MANAGEMENT ===")

	repo := &MockUserRepository{
		users: map[int]*User{
			1: {ID: 1, Name: "Alice"},
			2: {ID: 2, Name: "Bob"},
		},
	}

	// Test List() - doit retourner des copies
	users := repo.List()
	fmt.Printf("Original Alice: %p\n", repo.users[1])
	fmt.Printf("Liste Alice: %p\n", users[0])
	fmt.Printf("Même adresse? %t (doit être false pour sécurité)\n",
		repo.users[1] == users[0])

	// Modifier via la liste ne doit pas affecter l'original
	users[0].Name = "Modified Alice"
	fmt.Printf("Original après modification liste: %s (doit être inchangé)\n",
		repo.users[1].Name)

	// Test Update() - modifie en place
	updateUser := &User{ID: 1, Name: "Updated Alice"}
	oldAddr := repo.users[1]
	repo.Update(updateUser)

	fmt.Printf("Adresse avant update: %p\n", oldAddr)
	fmt.Printf("Adresse après update: %p\n", repo.users[1])
	fmt.Printf("Même adresse après update? %t (efficace si true)\n",
		oldAddr == repo.users[1])
}

func main() {
	demonstratePointerVsCopy()
	demonstrateList()
	demonstrateUpdate()
	demonstrateStackVsHeap()
	demonstrateRepositoryMemory()
}

// =============================================================================
// RÉSUMÉ DES BONNES PRATIQUES
// =============================================================================

/*
1. SLICE DE POINTEURS []*User :
   - Chaque élément pointe vers un User en mémoire
   - Si vous retournez les pointeurs originaux → modifications possibles
   - Si vous retournez des copies → sécurisé mais plus de mémoire

2. UPDATE avec *User :
   - Reçoit un pointeur vers l'objet à updater
   - Peut modifier en place (efficace) ou remplacer (simple mais coûteux)

3. RÈGLES DE SÉCURITÉ :
   - Repository.List() → retourner des copies
   - Repository.GetByID() → retourner une copie
   - Repository.Update() → modifier en place si possible

4. ALLOCATION MÉMOIRE :
   - Variables locales → stack (rapide)
   - new(), make(), variables qui "s'échappent" → heap
   - Go gère automatiquement (garbage collector)

5. PERFORMANCE :
   - Copies = plus de mémoire mais sécurisé
   - Pointeurs partagés = économique mais risqué
   - Choisir selon le contexte
*/
