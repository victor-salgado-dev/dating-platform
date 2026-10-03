package profiles

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// invalidField debe producir algo que writeProfileError reconozca como
// ValidationError (puntero vs. valor). Si este test falla, TODOS los errores
// de validación saldrían como 500.
func TestInvalidFieldIsValidationError(t *testing.T) {
	err := fmt.Errorf("envuelto: %w", invalidField("bio", "demasiado largo"))

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(*ValidationError) no reconoce el error de invalidField: %T", errors.Unwrap(err))
	}
}

func TestWriteProfileError_Status(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"validación", invalidField("bio", "x"), http.StatusBadRequest},
		{"perfil no encontrado", ErrNotFound, http.StatusNotFound},
		{"perfil ya existe", ErrAlreadyExists, http.StatusConflict},
		{"foto no encontrada", ErrPhotoNotFound, http.StatusNotFound},
		{"demasiadas fotos", ErrTooManyPhotos, http.StatusBadRequest},
		{"interés no encontrado", ErrInterestNotFound, http.StatusBadRequest},
		{"envuelto", fmt.Errorf("ctx: %w", ErrNotFound), http.StatusNotFound},
		{"desconocido", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeProfileError(rec, tc.err)
			if rec.Code != tc.want {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tc.want)
			}
		})
	}
}

func TestWritePublicProfileError_Status(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"perfil no encontrado", ErrNotFound, http.StatusNotFound},
		{"foto no encontrada", ErrPhotoNotFound, http.StatusNotFound}, // antes: 500
		{"desconocido", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writePublicProfileError(rec, tc.err)
			if rec.Code != tc.want {
				t.Errorf("status = %d, se esperaba %d", rec.Code, tc.want)
			}
		})
	}
}
