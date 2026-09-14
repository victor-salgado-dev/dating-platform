package consent

import "errors"

// ErrTermsNotAccepted: el registro exige marcar explícitamente que se
// han leído y aceptado los Términos y la Política de Privacidad.
var ErrTermsNotAccepted = errors.New("consent: debes aceptar los Términos y la Política de Privacidad")
