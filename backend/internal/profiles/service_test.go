package profiles

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestAgeAt(t *testing.T) {
	cases := []struct {
		name      string
		birthDate time.Time
		now       time.Time
		want      int
	}{
		{"cumpleaños ya pasado este año", date(2000, time.January, 1), date(2024, time.June, 15), 24},
		{"cumpleaños es justo hoy", date(2000, time.June, 15), date(2024, time.June, 15), 24},
		{"cumpleaños todavía no llega este año", date(2000, time.December, 25), date(2024, time.June, 15), 23},
		{"nacido el mismo día del año, un año antes", date(2023, time.June, 15), date(2024, time.June, 15), 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := AgeAt(tc.birthDate, tc.now)
			if got != tc.want {
				t.Errorf("AgeAt(%v, %v) = %d, se esperaba %d", tc.birthDate, tc.now, got, tc.want)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	if err := validateDisplayName("Ada"); err != nil {
		t.Errorf("nombre válido rechazado: %v", err)
	}
	if err := validateDisplayName("   "); err == nil {
		t.Error("nombre vacío (solo espacios) debería rechazarse")
	}
	tooLong := make([]byte, MaxDisplayNameLen+1)
	for i := range tooLong {
		tooLong[i] = 'a'
	}
	if err := validateDisplayName(string(tooLong)); err == nil {
		t.Error("nombre demasiado largo debería rechazarse")
	}
}

func TestValidateBirthDate(t *testing.T) {
	now := time.Now()

	adult := now.AddDate(-25, 0, 0)
	if err := validateBirthDate(adult); err != nil {
		t.Errorf("fecha de nacimiento de un adulto rechazada: %v", err)
	}

	minor := now.AddDate(-17, 0, 0)
	if err := validateBirthDate(minor); err == nil {
		t.Error("fecha de nacimiento de un menor de 18 debería rechazarse (sección 1: V1 solo mayores de edad)")
	}

	if err := validateBirthDate(time.Time{}); err == nil {
		t.Error("fecha de nacimiento vacía debería rechazarse")
	}
}

func TestValidateCountryCode(t *testing.T) {
	valid := []string{"ES", "US", "FR"}
	for _, v := range valid {
		if err := validateCountryCode(v); err != nil {
			t.Errorf("código de país válido %q rechazado: %v", v, err)
		}
	}

	invalid := []string{"es", "ESP", "E1", "", "E"}
	for _, v := range invalid {
		if err := validateCountryCode(v); err == nil {
			t.Errorf("código de país inválido %q debería rechazarse", v)
		}
	}
}

func TestIsValidGenderAndRelationshipGoal(t *testing.T) {
	if !IsValidGender(GenderFemale) || !IsValidGender(GenderMale) {
		t.Error("géneros de la lista permitida deberían ser válidos")
	}
	if IsValidGender(Gender("robot")) {
		t.Error("un género fuera de la lista permitida no debería ser válido")
	}

	if !IsValidRelationshipGoal(RelationshipCasual) {
		t.Error("objetivo de relación de la lista permitida debería ser válido")
	}
	if IsValidRelationshipGoal(RelationshipGoal("marriage-ish")) {
		t.Error("un objetivo de relación fuera de la lista no debería ser válido")
	}
}

func TestValidateBio(t *testing.T) {
	if err := validateBio(nil); err != nil {
		t.Errorf("bio ausente (nil) no debería ser un error: %v", err)
	}

	short := "Me encanta viajar."
	if err := validateBio(&short); err != nil {
		t.Errorf("bio corta válida rechazada: %v", err)
	}

	tooLongRunes := make([]rune, MaxBioLen+1)
	for i := range tooLongRunes {
		tooLongRunes[i] = 'a'
	}
	tooLong := string(tooLongRunes)
	if err := validateBio(&tooLong); err == nil {
		t.Error("bio demasiado larga debería rechazarse")
	}
}

func TestValidateInterests(t *testing.T) {
	if err := validateInterests([]string{"ajedrez", "senderismo"}); err != nil {
		t.Errorf("intereses válidos rechazados: %v", err)
	}
	if err := validateInterests([]string{"ajedrez", ""}); err == nil {
		t.Error("un interés vacío debería rechazarse")
	}

	tooMany := make([]string, MaxInterests+1)
	for i := range tooMany {
		tooMany[i] = "interes"
	}
	if err := validateInterests(tooMany); err == nil {
		t.Error("demasiados intereses deberían rechazarse")
	}
}

func TestValidateLanguages(t *testing.T) {
	if err := validateLanguages([]string{"es", "en"}); err != nil {
		t.Errorf("idiomas válidos rechazados: %v", err)
	}
	if err := validateLanguages([]string{"es", " "}); err == nil {
		t.Error("un idioma vacío debería rechazarse")
	}
}
