package search

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

func TestToSearchResponse_PhotoURLs(t *testing.T) {
	profileID, photoID := uuid.New(), uuid.New()
	res := &Result{Items: []ResultItem{
		{ProfileID: profileID, PhotoID: &photoID, HasPhoto: true, CreatedAt: time.Unix(0, 0).UTC()},
		{ProfileID: uuid.New(), CreatedAt: time.Unix(0, 0).UTC()}, // sin foto
	}}

	resp := toSearchResponse(res)

	with, without := resp.Items[0], resp.Items[1]
	if with.PhotoURL == nil || *with.PhotoURL != profiles.PublicPhotoURL(profileID, photoID, false) {
		t.Errorf("photo_url = %v", with.PhotoURL)
	}
	if with.PhotoThumbURL == nil || *with.PhotoThumbURL != profiles.PublicPhotoURL(profileID, photoID, true) {
		t.Errorf("photo_thumb_url = %v", with.PhotoThumbURL)
	}
	if !strings.HasSuffix(*with.PhotoThumbURL, "?size=thumb") || strings.Contains(*with.PhotoURL, "size=") {
		t.Errorf("la miniatura debe llevar ?size=thumb y la foto completa no: %s / %s", *with.PhotoURL, *with.PhotoThumbURL)
	}
	if without.PhotoURL != nil || without.PhotoThumbURL != nil {
		t.Error("un perfil sin foto no tiene ninguna URL de foto")
	}

	// Contrato JSON: photo_url no cambia y photo_thumb_url es un campo nuevo (aditivo).
	raw, _ := json.Marshal(without)
	for _, key := range []string{`"photo_url":null`, `"photo_thumb_url":null`, `"relationship_goals":[]`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("falta %s en el JSON: %s", key, raw)
		}
	}
}
