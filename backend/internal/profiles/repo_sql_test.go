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

// VisibleSQL es la regla de visibilidad que reutilizan otros paquetes: debe ser
// exactamente visiblePredicate con el placeholder cambiado.
func TestVisibleSQL(t *testing.T) {
	if got := VisibleSQL("$2"); got != visiblePredicate {
		t.Errorf("con $2 debe devolver visiblePredicate tal cual:\n%s", got)
	}

	sql := VisibleSQL("$1")
	if strings.Contains(sql, "$2") {
		t.Errorf("no debe quedar ningún $2:\n%s", sql)
	}
	if n := strings.Count(sql, "$1"); n != 2 {
		t.Errorf("el viewer aparece en los dos sentidos del bloqueo (2 veces), aparece %d:\n%s", n, sql)
	}
	for _, want := range []string{"u.status = 'active'", "u.deleted_at IS NULL",
		"b.blocker_id = $1 AND b.blocked_id = p.user_id", "b.blocker_id = p.user_id AND b.blocked_id = $1"} {
		if !strings.Contains(sql, want) {
			t.Errorf("falta %q en:\n%s", want, sql)
		}
	}
	if got := VisibleSQL("$12"); !strings.Contains(got, "$12") || strings.Contains(got, "$2 ") {
		t.Errorf("placeholder de varios dígitos mal sustituido:\n%s", got)
	}
}

func TestVisibleSQL_RejectsAnythingButAPlaceholder(t *testing.T) {
	for _, bad := range []string{"", "$", "1", "$0", "$x", "$1; DROP TABLE users", "$1 OR 1=1", "@1", "$-1", " $1"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("VisibleSQL(%q) debería entrar en pánico", bad)
				}
			}()
			VisibleSQL(bad)
		}()
	}
}
