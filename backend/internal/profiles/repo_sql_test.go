package profiles

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// squish colapsa los espacios para comparar SQL sin depender del formato.
func squish(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestBuildProfileUpdate(t *testing.T) {
	userID := uuid.New()
	bio := "hola"

	query, args, ok := buildProfileUpdate(userID, ProfilePatch{
		Region: FieldOf[*string](nil), // null: borrar
		Bio:    FieldOf(&bio),
	})
	if !ok {
		t.Fatal("ok=false con columnas que actualizar")
	}
	want := "UPDATE profiles SET region = $1, bio = $2, updated_at = now() WHERE user_id = $3"
	if squish(query) != want {
		t.Errorf("SQL:\n got: %s\nwant: %s", squish(query), want)
	}
	if len(args) != 3 || args[0] != (*string)(nil) || args[1] != &bio || args[2] != userID {
		t.Errorf("args = %#v", args)
	}

	if _, _, ok := buildProfileUpdate(userID, ProfilePatch{}); ok {
		t.Error("un patch vacío no debe generar UPDATE")
	}
}

func TestBuildPartnerPreferencesUpsert(t *testing.T) {
	profileID := uuid.New()
	min, fun := 30, 5

	query, args := buildPartnerPreferencesUpsert(profileID, PartnerPreferencesPatch{
		AgeMin:        FieldOf(&min),
		ImportanceFun: FieldOf(&fun),
	})
	want := "INSERT INTO profile_partner_preferences (profile_id, age_min, importance_fun) " +
		"VALUES ($1, $2, $3) " +
		"ON CONFLICT (profile_id) DO UPDATE SET age_min = EXCLUDED.age_min, importance_fun = EXCLUDED.importance_fun, updated_at = now()"
	if squish(query) != want {
		t.Errorf("SQL:\n got: %s\nwant: %s", squish(query), want)
	}
	if !reflect.DeepEqual(args, []any{profileID, &min, &fun}) {
		t.Errorf("args = %#v", args)
	}

	// Sin claves: solo asegura que exista la fila.
	query, args = buildPartnerPreferencesUpsert(profileID, PartnerPreferencesPatch{})
	want = "INSERT INTO profile_partner_preferences (profile_id) VALUES ($1) " +
		"ON CONFLICT (profile_id) DO UPDATE SET updated_at = profile_partner_preferences.updated_at"
	if squish(query) != want || len(args) != 1 {
		t.Errorf("patch vacío:\n got: %s (%d args)\nwant: %s", squish(query), len(args), want)
	}
}
