package profiles

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func rawJSON(t *testing.T, s string) map[string]json.RawMessage {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		t.Fatalf("JSON de prueba inválido: %v", err)
	}
	return raw
}

// Las tres situaciones que un puntero simple no distingue.
func TestField_AbsentNullAndValue(t *testing.T) {
	var p ProfilePatch
	body := `{"region": null, "bio": "hola", "body_art": [], "sports": null, "height": 180, "display_name": null}`
	if err := decodePatchKeys(rawJSON(t, body), &p); err != nil {
		t.Fatalf("decodePatchKeys: %v", err)
	}

	if !p.Region.Set || p.Region.Value != nil {
		t.Errorf("region:null debe ser Set con valor nil, es %+v", p.Region)
	}
	if !p.Bio.Set || p.Bio.Value == nil || *p.Bio.Value != "hola" {
		t.Errorf("bio mal decodificado: %+v", p.Bio)
	}
	if !p.BodyArt.Set || p.BodyArt.Value == nil || len(p.BodyArt.Value) != 0 {
		t.Errorf("body_art:[] debe ser Set con slice vacío NO nil, es %#v", p.BodyArt)
	}
	if !p.Sports.Set || p.Sports.Value != nil {
		t.Errorf("sports:null debe ser Set con slice nil, es %#v", p.Sports)
	}
	if !p.Height.Set || p.Height.Value == nil || *p.Height.Value != 180 {
		t.Errorf("height mal decodificado: %+v", p.Height)
	}
	if !p.DisplayName.Set || p.DisplayName.Value != "" {
		t.Errorf("display_name:null debe ser Set con valor vacío (lo rechaza la validación), es %+v", p.DisplayName)
	}
	if p.Weight.Set || p.Gender.Set || p.CountryCode.Set {
		t.Errorf("las claves ausentes no deben quedar Set")
	}
}

func TestField_UnknownKeysAreIgnored(t *testing.T) {
	var p ProfilePatch
	// Un cliente que reenvía el perfil entero: id, age, created_at no existen en el patch.
	if err := decodePatchKeys(rawJSON(t, `{"id":"x","age":33,"created_at":"2020-01-01","bio":"ok"}`), &p); err != nil {
		t.Fatalf("las claves desconocidas no deben fallar (rejectUnknownPatchFields=%v): %v", rejectUnknownPatchFields, err)
	}
	if !p.Bio.Set {
		t.Error("bio debería haberse decodificado")
	}
}

func TestField_BirthDate(t *testing.T) {
	var p ProfilePatch
	if err := decodePatchKeys(rawJSON(t, `{"birth_date":"1990-05-17"}`), &p); err != nil {
		t.Fatal(err)
	}
	want := time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC)
	if !p.BirthDate.Set || !p.BirthDate.Value.Time().Equal(want) {
		t.Errorf("birth_date = %v, se esperaba %v", p.BirthDate.Value.Time(), want)
	}
}

// Campo y mensaje que recibe el usuario ante un tipo incorrecto: los mismos
// textos que daba el antiguo buildProfilePatch.
func TestField_TypeErrors(t *testing.T) {
	cases := []struct {
		name, body, field, msg string
		partner                bool
	}{
		{"texto obligatorio", `{"display_name": 5}`, "display_name", "debe ser texto", false},
		{"texto opcional", `{"bio": 5}`, "bio", "debe ser texto o null", false},
		{"entero", `{"height": "alto"}`, "height", "debe ser un entero o null", false},
		{"entero con decimales", `{"height": 1.5}`, "height", "debe ser un entero o null", false},
		{"lista", `{"sports": "futbol"}`, "sports", "debe ser una lista de textos o null", false},
		{"lista de números", `{"sports": [1, 2]}`, "sports", "debe ser una lista de textos o null", false},
		{"objetivos", `{"relationship_goals": 5}`, "relationship_goals", "debe ser una lista de textos o null", false},
		{"fecha mal formada", `{"birth_date": "01/02/1990"}`, "birth_date", "formato esperado YYYY-MM-DD", false},
		{"fecha como número", `{"birth_date": 19900101}`, "birth_date", "formato esperado YYYY-MM-DD", false},
		{"fecha null", `{"birth_date": null}`, "birth_date", "formato esperado YYYY-MM-DD", false},
		{"preferencias: entero", `{"age_min": "x"}`, "age_min", "debe ser un entero o null", true},
		{"preferencias: lista", `{"desired_traits": 3}`, "desired_traits", "debe ser una lista de textos o null", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var dst any = &ProfilePatch{}
			if tc.partner {
				dst = &PartnerPreferencesPatch{}
			}
			err := decodePatchKeys(rawJSON(t, tc.body), dst)
			if err == nil {
				t.Fatal("se esperaba un error")
			}
			field, msg := describePatchError(err)
			if field != tc.field || msg != tc.msg {
				t.Errorf("campo=%q mensaje=%q, se esperaba %q / %q", field, msg, tc.field, tc.msg)
			}
			if unmarshalPatch(rawJSON(t, tc.body), dst) == nil {
				t.Error("unmarshalPatch debe devolver error")
			}
		})
	}
}

func TestSetColumns_OnlySetFieldsAndDBValues(t *testing.T) {
	bio := "hola"
	birth := time.Date(1990, 5, 17, 0, 0, 0, 0, time.UTC)
	patch := ProfilePatch{
		DisplayName:       FieldOf("Ana"),
		BirthDate:         FieldOf(DateOnly(birth)),
		Gender:            FieldOf(GenderFemale),
		Region:            FieldOf[*string](nil), // null: borrar
		RelationshipGoals: FieldOf([]RelationshipGoal{RelationshipLongTerm}),
		Bio:               FieldOf(&bio),
	}

	cols, vals := setColumns(patch)

	wantCols := []string{"display_name", "birth_date", "gender", "region", "relationship_goals", "bio"}
	if !reflect.DeepEqual(cols, wantCols) {
		t.Fatalf("columnas = %v, se esperaba %v", cols, wantCols)
	}
	if v, ok := vals[0].(string); !ok || v != "Ana" {
		t.Errorf("display_name: %#v", vals[0])
	}
	if v, ok := vals[1].(time.Time); !ok || !v.Equal(birth) {
		t.Errorf("birth_date debe enviarse como time.Time: %#v", vals[1])
	}
	if v, ok := vals[2].(string); !ok || v != "female" {
		t.Errorf("gender debe enviarse como string: %#v", vals[2])
	}
	if v, ok := vals[3].(*string); !ok || v != nil {
		t.Errorf("region null debe enviarse como (*string)(nil): %#v", vals[3])
	}
	if v, ok := vals[4].([]string); !ok || len(v) != 1 || v[0] != "long_term" {
		t.Errorf("relationship_goals debe enviarse como []string: %#v", vals[4])
	}
	if v, ok := vals[5].(*string); !ok || v == nil || *v != "hola" {
		t.Errorf("bio: %#v", vals[5])
	}

	if cols, _ := setColumns(ProfilePatch{}); len(cols) != 0 {
		t.Errorf("un patch vacío no debe producir columnas, produjo %v", cols)
	}
}

// --- Contrato: etiquetas json == columnas == DTOs de respuesta -------------------

func jsonNames(t reflect.Type) []string {
	var names []string
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name != "" && name != "-" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func without(names []string, drop ...string) []string {
	var out []string
	for _, n := range names {
		skip := false
		for _, d := range drop {
			if n == d {
				skip = true
			}
		}
		if !skip {
			out = append(out, n)
		}
	}
	return out
}

func TestProfilePatch_CoversExactlyTheProfileColumns(t *testing.T) {
	var cols []string
	for _, c := range strings.Split(allProfileCols, ",") {
		if c = strings.TrimSpace(c); c != "" && c != "user_id" {
			cols = append(cols, c)
		}
	}
	sort.Strings(cols)

	got := jsonNames(reflect.TypeOf(ProfilePatch{}))
	if !reflect.DeepEqual(got, cols) {
		t.Errorf("ProfilePatch no coincide con allProfileCols\n patch: %v\n  cols: %v", got, cols)
	}
}

func TestProfileFieldsStayInSync(t *testing.T) {
	required := []string{"display_name", "birth_date", "gender", "country_code"}

	details := jsonNames(reflect.TypeOf(ProfileDetails{}))
	patch := jsonNames(reflect.TypeOf(ProfilePatch{}))
	response := jsonNames(reflect.TypeOf(profileResponse{}))

	wantPatch := append(append([]string{}, details...), required...)
	sort.Strings(wantPatch)
	if !reflect.DeepEqual(patch, wantPatch) {
		t.Errorf("ProfilePatch debe ser ProfileDetails + los 4 obligatorios\n patch: %v\n  want: %v", patch, wantPatch)
	}

	// Todo lo que se puede escribir debe poder leerse en la respuesta, salvo la
	// fecha de nacimiento: la API expone la edad (age), no el dato en sí.
	wantResponse := append(without(wantPatch, "birth_date"), "id", "age", "created_at", "updated_at")
	sort.Strings(wantResponse)
	if !reflect.DeepEqual(response, wantResponse) {
		t.Errorf("profileResponse no coincide con los campos del perfil\n response: %v\n     want: %v", response, wantResponse)
	}
}

func TestPartnerPreferencesFieldsStayInSync(t *testing.T) {
	patch := jsonNames(reflect.TypeOf(PartnerPreferencesPatch{}))
	response := without(jsonNames(reflect.TypeOf(partnerPreferencesResponse{})), "updated_at")
	if !reflect.DeepEqual(patch, response) {
		t.Errorf("PartnerPreferencesPatch no coincide con partnerPreferencesResponse\n patch: %v\n resp:  %v", patch, response)
	}
}

// Un campo del patch que NO sea Field[T] se ignoraría en silencio al construir el UPDATE.
func TestPatches_AllFieldsAreFields(t *testing.T) {
	iface := reflect.TypeOf((*patchColumn)(nil)).Elem()
	for _, typ := range []reflect.Type{reflect.TypeOf(ProfilePatch{}), reflect.TypeOf(PartnerPreferencesPatch{})} {
		for i := 0; i < typ.NumField(); i++ {
			if f := typ.Field(i); !f.Type.Implements(iface) {
				t.Errorf("%s.%s no es un Field[T]: setColumns lo ignoraría", typ.Name(), f.Name)
			}
		}
	}
}

// El cuerpo de POST /profiles se decodifica directamente en createProfileRequest.
func TestCreateProfileRequest_DecodesEmbeddedDetails(t *testing.T) {
	body := `{"display_name":"Ana","birth_date":"1990-05-17","gender":"female","country_code":"es",
		"region":"Madrid","relationship_goals":["long_term","friendship"],"height":170,"sports":["yoga"],
		"id":"ignorado"}`
	var req createProfileRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	if req.DisplayName != "Ana" || req.Gender != GenderFemale || req.BirthDate != "1990-05-17" {
		t.Errorf("campos obligatorios mal decodificados: %+v", req)
	}
	if req.Region == nil || *req.Region != "Madrid" || req.Height == nil || *req.Height != 170 {
		t.Errorf("campos de ProfileDetails mal decodificados: %+v", req.ProfileDetails)
	}
	if len(req.RelationshipGoals) != 2 || req.RelationshipGoals[0] != RelationshipLongTerm {
		t.Errorf("relationship_goals = %v", req.RelationshipGoals)
	}
	if req.Weight != nil || req.BodyArt != nil {
		t.Error("los campos ausentes deben quedar nil")
	}
}
