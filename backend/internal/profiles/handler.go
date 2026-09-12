package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
)

const dateLayout = "2006-01-02"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs -----------------------------------------------------------------

type profileResponse struct {
	ID               string   `json:"id"`
	DisplayName      string   `json:"display_name"`
	Age              int      `json:"age"`
	Gender           string   `json:"gender"`
	CountryCode      string   `json:"country_code"`
	Region           *string  `json:"region"`
	Languages        []string `json:"languages"`
	RelationshipGoal *string  `json:"relationship_goal"`
	HasChildren      *bool    `json:"has_children"`
	WantsChildren    *bool    `json:"wants_children"`
	Bio              *string  `json:"bio"`
	Interests        []string `json:"interests"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

func toProfileResponse(p *Profile) profileResponse {
	var relGoal *string
	if p.RelationshipGoal != nil {
		v := string(*p.RelationshipGoal)
		relGoal = &v
	}
	return profileResponse{
		ID:               p.ID.String(),
		DisplayName:      p.DisplayName,
		Age:              p.Age(),
		Gender:           string(p.Gender),
		CountryCode:      p.CountryCode,
		Region:           p.Region,
		Languages:        p.Languages,
		RelationshipGoal: relGoal,
		HasChildren:      p.HasChildren,
		WantsChildren:    p.WantsChildren,
		Bio:              p.Bio,
		Interests:        p.Interests,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
	}
}

type photoResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

func toPhotoResponseSelf(ph *Photo) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       fmt.Sprintf("/api/v1/profiles/me/photos/%s/file", ph.ID.String()),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

func toPhotoResponsePublic(ph *Photo, profileID uuid.UUID) photoResponse {
	return photoResponse{
		ID:        ph.ID.String(),
		URL:       fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file", profileID.String(), ph.ID.String()),
		Position:  ph.Position,
		CreatedAt: ph.CreatedAt.Format(time.RFC3339),
	}
}

type createProfileRequest struct {
	DisplayName      string   `json:"display_name"`
	BirthDate        string   `json:"birth_date"`
	Gender           string   `json:"gender"`
	CountryCode      string   `json:"country_code"`
	Region           *string  `json:"region"`
	Languages        []string `json:"languages"`
	RelationshipGoal *string  `json:"relationship_goal"`
	HasChildren      *bool    `json:"has_children"`
	WantsChildren    *bool    `json:"wants_children"`
	Bio              *string  `json:"bio"`
	Interests        []string `json:"interests"`
}

// --- Handlers: perfil -------------------------------------------------

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	p, err := h.svc.GetMyProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Todavía no has creado tu perfil.")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo obtener el perfil.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	var req createProfileRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	birthDate, err := time.Parse(dateLayout, req.BirthDate)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", "birth_date debe tener formato YYYY-MM-DD.")
		return
	}

	var relGoal *RelationshipGoal
	if req.RelationshipGoal != nil {
		g := RelationshipGoal(*req.RelationshipGoal)
		relGoal = &g
	}

	p, err := h.svc.CreateProfile(r.Context(), userID, CreateProfileInput{
		DisplayName:      req.DisplayName,
		BirthDate:        birthDate,
		Gender:           Gender(req.Gender),
		CountryCode:      req.CountryCode,
		Region:           req.Region,
		Languages:        req.Languages,
		RelationshipGoal: relGoal,
		HasChildren:      req.HasChildren,
		WantsChildren:    req.WantsChildren,
		Bio:              req.Bio,
		Interests:        req.Interests,
	})
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toProfileResponse(p))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	var raw map[string]json.RawMessage
	if err := httpx.DecodeJSON(r, &raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return
	}

	patch, err := buildProfilePatch(raw)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	p, err := h.svc.UpdateProfile(r.Context(), userID, patch)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

// buildProfilePatch traduce el JSON de la petición a un ProfilePatch,
// distinguiendo "clave ausente" (no tocar) de "clave con valor null"
// (borrar, pasa a desconocido) tal como exige la regla de datos
// faltantes para los campos opcionales.
func buildProfilePatch(raw map[string]json.RawMessage) (ProfilePatch, error) {
	var patch ProfilePatch

	if v, ok := raw["display_name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("display_name", "debe ser texto")
		}
		patch.DisplayName = &s
	}

	if v, ok := raw["birth_date"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("birth_date", "debe ser texto")
		}
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			return patch, invalidField("birth_date", "formato esperado YYYY-MM-DD")
		}
		patch.BirthDate = &t
	}

	if v, ok := raw["gender"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("gender", "debe ser texto")
		}
		g := Gender(s)
		patch.Gender = &g
	}

	if v, ok := raw["country_code"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return patch, invalidField("country_code", "debe ser texto")
		}
		patch.CountryCode = &s
	}

	if v, ok := raw["region"]; ok {
		patch.RegionSet = true
		if !isJSONNull(v) {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("region", "debe ser texto o null")
			}
			patch.Region = &s
		}
	}

	if v, ok := raw["languages"]; ok {
		patch.LanguagesSet = true
		if !isJSONNull(v) {
			var s []string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("languages", "debe ser una lista de textos o null")
			}
			patch.Languages = s
		}
	}

	if v, ok := raw["relationship_goal"]; ok {
		patch.RelationshipGoalSet = true
		if !isJSONNull(v) {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("relationship_goal", "debe ser texto o null")
			}
			g := RelationshipGoal(s)
			patch.RelationshipGoal = &g
		}
	}

	if v, ok := raw["has_children"]; ok {
		patch.HasChildrenSet = true
		if !isJSONNull(v) {
			var b bool
			if err := json.Unmarshal(v, &b); err != nil {
				return patch, invalidField("has_children", "debe ser booleano o null")
			}
			patch.HasChildren = &b
		}
	}

	if v, ok := raw["wants_children"]; ok {
		patch.WantsChildrenSet = true
		if !isJSONNull(v) {
			var b bool
			if err := json.Unmarshal(v, &b); err != nil {
				return patch, invalidField("wants_children", "debe ser booleano o null")
			}
			patch.WantsChildren = &b
		}
	}

	if v, ok := raw["bio"]; ok {
		patch.BioSet = true
		if !isJSONNull(v) {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("bio", "debe ser texto o null")
			}
			patch.Bio = &s
		}
	}

	if v, ok := raw["interests"]; ok {
		patch.InterestsSet = true
		if !isJSONNull(v) {
			var s []string
			if err := json.Unmarshal(v, &s); err != nil {
				return patch, invalidField("interests", "debe ser una lista de textos o null")
			}
			patch.Interests = s
		}
	}

	return patch, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return string(raw) == "null"
}

// --- Handlers: fotos ----------------------------------------------------

func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxPhotoSizeBytes+1<<20) // margen para el multipart
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "No se pudo leer el fichero enviado (¿supera el límite de tamaño?).")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "Falta el fichero 'photo' en el formulario.")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	photo, err := h.svc.UploadPhoto(r.Context(), userID, contentType, header.Size, file)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toPhotoResponseSelf(photo))
}

func (h *Handler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	photos, err := h.svc.ListPhotos(r.Context(), userID)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	resp := make([]photoResponse, 0, len(photos))
	for i := range photos {
		resp = append(resp, toPhotoResponseSelf(&photos[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ServePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	photoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	rc, ph, err := h.svc.OpenPhoto(r.Context(), userID, photoID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", ph.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = io.Copy(w, rc)
}

func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	photoID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	if err := h.svc.DeletePhoto(r.Context(), userID, photoID); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Handlers: perfiles públicos (Fase 6) --------------------------------
//
// A diferencia de los handlers /me, estos actúan sobre el perfil de OTRA
// persona identificado por {profileID} en la ruta. La privacidad básica
// de esta fase consiste en que Service.GetPublicProfile (y las llamadas
// que dependen de ella) solo devuelven perfiles de cuentas activas: ver
// el perfil de alguien suspendido o que se ha dado de baja da el mismo
// 404 genérico que un ID inexistente, para no filtrar esa información.

func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}

	p, err := h.svc.GetPublicProfile(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) ListPublicPhotos(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}

	photos, err := h.svc.ListPublicPhotos(r.Context(), viewerID, profileID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}

	resp := make([]photoResponse, 0, len(photos))
	for i := range photos {
		resp = append(resp, toPhotoResponsePublic(&photos[i], profileID))
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) ServePublicPhoto(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
		return
	}

	profileID, err := uuid.Parse(r.PathValue("profileID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de perfil inválido.")
		return
	}
	photoID, err := uuid.Parse(r.PathValue("photoID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de foto inválido.")
		return
	}

	rc, ph, err := h.svc.OpenPublicPhoto(r.Context(), viewerID, profileID, photoID)
	if err != nil {
		writePublicProfileError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", ph.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = io.Copy(w, rc)
}

// --- Errores --------------------------------------------------------------

func writePublicProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Perfil no encontrado.")
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}

func writeProfileError(w http.ResponseWriter, err error) {
	var valErr *ValidationError
	switch {
	case errors.As(err, &valErr):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_field", valErr.Error())
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "profile_not_found", "Todavía no has creado tu perfil.")
	case errors.Is(err, ErrAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "profile_already_exists", "Ya tienes un perfil creado.")
	case errors.Is(err, ErrPhotoNotFound):
		httpx.WriteError(w, http.StatusNotFound, "photo_not_found", "Foto no encontrada.")
	case errors.Is(err, ErrTooManyPhotos):
		httpx.WriteError(w, http.StatusConflict, "too_many_photos", fmt.Sprintf("Máximo %d fotos por perfil.", MaxPhotosPerProfile))
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.")
	}
}
