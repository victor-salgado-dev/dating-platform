package profiles

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"
)

// Field es un campo opcional de un PATCH: distingue "no viene en el body" de
// "viene a null", algo que un simple puntero no puede expresar.
//
//	clave ausente  -> Set=false (no tocar)
//	clave a null   -> Set=true, Value = valor cero (nil, "", 0): borrar el dato
//	clave con dato -> Set=true, Value = el dato
//
// Funciona porque encoding/json solo llama a UnmarshalJSON cuando la clave
// existe en el JSON (incluido null). Para campos que admiten "desconocido" usa
// T = *string, *int o []string; para obligatorios, T = string, etc.
type Field[T any] struct {
	Set   bool
	Value T
}

// FieldOf devuelve un Field presente (Set=true) con el valor v: la forma corta
// de construir un patch en código (tests, scripts).
func FieldOf[T any](v T) Field[T] { return Field[T]{Set: true, Value: v} }

// UnmarshalJSON marca el campo como presente y decodifica el valor. Un tipo
// incorrecto se devuelve como *json.UnmarshalTypeError con Type = T: el
// decodificador le añade el nombre de la clave (Field) y así puede traducirse
// a un invalidField con el campo exacto (ver patchDecodeError).
func (f *Field[T]) UnmarshalJSON(b []byte) error {
	f.Set = true
	if err := json.Unmarshal(b, &f.Value); err != nil {
		var ute *json.UnmarshalTypeError
		if errors.As(err, &ute) {
			return &json.UnmarshalTypeError{Value: ute.Value, Type: reflect.TypeOf((*T)(nil)).Elem(), Offset: ute.Offset}
		}
		return err
	}
	return nil
}

// patchColumn lo implementa todo Field[T]; permite recorrer un patch sin
// conocer el tipo concreto de cada campo.
type patchColumn interface {
	isSet() bool
	dbValue() any
}

func (f Field[T]) isSet() bool { return f.Set }

// dbValue convierte el valor al tipo que se envía a Postgres. Los tipos
// propios del dominio se pasan a su tipo base; el resto va tal cual.
func (f Field[T]) dbValue() any {
	switch v := any(f.Value).(type) {
	case Gender:
		return string(v)
	case DateOnly:
		return v.Time()
	case []RelationshipGoal:
		return relationshipGoalsToDB(v)
	}
	return f.Value
}

// DateOnly es una fecha civil sin hora ni zona, en JSON "YYYY-MM-DD".
type DateOnly time.Time

// Time devuelve la fecha como time.Time (medianoche UTC).
func (d DateOnly) Time() time.Time { return time.Time(d) }

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	badFormat := &json.UnmarshalTypeError{Value: "string", Type: reflect.TypeOf(DateOnly{})}

	var s string
	if err := json.Unmarshal(b, &s); err != nil || bytes.Equal(b, []byte("null")) {
		return badFormat
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return badFormat
	}
	*d = DateOnly(t)
	return nil
}

// setColumns recorre un patch (un struct cuyos campos son Field[T]) y devuelve,
// en orden de declaración, las columnas con Set=true y sus valores listos para
// pasar como argumentos de la consulta.
//
// El nombre de columna sale de la etiqueta db o, si no hay, de la json. Son
// constantes del código, nunca entrada del usuario, así que es seguro
// interpolarlas en el SQL (los valores siempre van parametrizados).
func setColumns(patch any) (cols []string, vals []any) {
	v := reflect.ValueOf(patch)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		pc, ok := v.Field(i).Interface().(patchColumn)
		if !ok || !pc.isSet() {
			continue
		}
		col := patchColumnName(sf)
		if col == "" {
			continue
		}
		cols = append(cols, col)
		vals = append(vals, pc.dbValue())
	}
	return cols, vals
}

// patchColumnName devuelve la columna de un campo del patch ("" si no tiene).
func patchColumnName(sf reflect.StructField) string {
	if col := sf.Tag.Get("db"); col != "" {
		return col
	}
	name, _, _ := strings.Cut(sf.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}
	return name
}

// rejectUnknownPatchFields hace que un PATCH con claves que no existen en el
// struct (un typo, o un cliente que reenvía el perfil entero con "id", "age"...)
// devuelva 400 en vez de ignorarlas. Se deja en false para no cambiar el
// comportamiento actual; ponlo en true cuando el cliente esté limpio.
const rejectUnknownPatchFields = false

// unmarshalPatch decodifica las claves de un PATCH en dst (un *ProfilePatch o
// *PartnerPreferencesPatch). Los errores de tipo se devuelven como invalidField
// con el nombre exacto de la clave.
func unmarshalPatch(raw map[string]json.RawMessage, dst any) error {
	if err := decodePatchKeys(raw, dst); err != nil {
		field, msg := describePatchError(err)
		return invalidField(field, msg)
	}
	return nil
}

// decodePatchKeys hace la decodificación en sí y devuelve el error de
// encoding/json tal cual (separado de unmarshalPatch para poder probarlo sin
// depender del formato de ValidationError).
func decodePatchKeys(raw map[string]json.RawMessage, dst any) error {
	body, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	if rejectUnknownPatchFields {
		dec.DisallowUnknownFields()
	}
	return dec.Decode(dst)
}

// describePatchError traduce un error de decodificación al campo afectado y a
// un mensaje para el usuario.
func describePatchError(err error) (field, msg string) {
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) && ute.Field != "" {
		return ute.Field, expectedTypeMessage(ute.Type)
	}
	return "body", "El cuerpo de la petición no es válido."
}

// expectedTypeMessage describe el tipo esperado con los mismos textos que
// usaba el antiguo buildProfilePatch ("debe ser texto o null"...).
func expectedTypeMessage(t reflect.Type) string {
	if t == reflect.TypeOf(DateOnly{}) {
		return "formato esperado YYYY-MM-DD"
	}

	base := t
	if base.Kind() == reflect.Ptr {
		base = base.Elem()
	}

	var noun string
	switch base.Kind() {
	case reflect.Slice:
		noun = "una lista de textos"
	case reflect.String:
		noun = "texto"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		noun = "un entero"
	default:
		noun = "un valor válido"
	}

	msg := "debe ser " + noun
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		msg += " o null"
	}
	return msg
}
