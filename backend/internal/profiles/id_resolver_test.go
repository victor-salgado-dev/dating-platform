package profiles

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func newTestResolver(max int) *IDResolver {
	return &IDResolver{cache: make(map[uuid.UUID]uuid.UUID), maxEntries: max}
}

// Al llenarse expulsa UNA entrada; antes vaciaba el mapa entero y todas las
// peticiones siguientes iban a Postgres a la vez.
func TestIDResolver_EvictsOneEntryWhenFull(t *testing.T) {
	r := newTestResolver(100)

	var last uuid.UUID
	for i := 0; i < 250; i++ {
		last = uuid.New()
		r.remember(uuid.New(), last)
	}
	if got := len(r.cache); got != 100 {
		t.Fatalf("el mapa debe quedarse EXACTAMENTE en el máximo (100), tiene %d", got)
	}

	// Lo último que se recordó está siempre presente.
	user := uuid.New()
	r.remember(user, last)
	if id, ok := r.cache[user]; !ok || id != last {
		t.Error("la entrada recién guardada debe estar en la caché")
	}
	if len(r.cache) != 100 {
		t.Errorf("tras insertar sigue acotado a 100, tiene %d", len(r.cache))
	}
}

func TestIDResolver_UpdatingExistingEntryDoesNotEvict(t *testing.T) {
	r := newTestResolver(3)
	users := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, u := range users {
		r.remember(u, uuid.New())
	}

	newID := uuid.New()
	r.remember(users[1], newID) // ya existía: no hay expulsión

	if len(r.cache) != 3 {
		t.Fatalf("tamaño = %d, se esperaba 3", len(r.cache))
	}
	for _, u := range users {
		if _, ok := r.cache[u]; !ok {
			t.Error("actualizar una entrada existente no debe expulsar ninguna")
		}
	}
	if r.cache[users[1]] != newID {
		t.Error("la entrada debe quedar actualizada")
	}
}

// El camino de acierto no toca la base de datos (db es nil aquí).
func TestIDResolver_ProfileIDHitSkipsDatabaseAndIsRaceFree(t *testing.T) {
	r := newTestResolver(1000)
	userID, profileID := uuid.New(), uuid.New()
	r.remember(userID, profileID)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if got, err := r.ProfileID(context.Background(), userID); err != nil || got != profileID {
				t.Errorf("ProfileID = %v, %v", got, err)
			}
		}()
		go func() { // escrituras concurrentes (provocan expulsiones en otras claves)
			defer wg.Done()
			r.remember(uuid.New(), uuid.New())
		}()
	}
	wg.Wait()
}
