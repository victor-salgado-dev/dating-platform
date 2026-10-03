package profiles

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhotoURLs(t *testing.T) {
	profileID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	photoID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	cases := []struct{ got, want string }{
		{ownPhotoURL(photoID, false), "/api/v1/profiles/me/photos/22222222-2222-2222-2222-222222222222/file"},
		{ownPhotoURL(photoID, true), "/api/v1/profiles/me/photos/22222222-2222-2222-2222-222222222222/file?size=thumb"},
		{publicPhotoURL(profileID, photoID, false), "/api/v1/profiles/11111111-1111-1111-1111-111111111111/photos/22222222-2222-2222-2222-222222222222/file"},
		{publicPhotoURL(profileID, photoID, true), "/api/v1/profiles/11111111-1111-1111-1111-111111111111/photos/22222222-2222-2222-2222-222222222222/file?size=thumb"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("URL = %s\n  want %s", tc.got, tc.want)
		}
	}
}

// El sufijo que genera la URL debe ser el que reconoce el handler.
func TestPhotoURLs_ThumbRoundTrip(t *testing.T) {
	profileID, photoID := uuid.New(), uuid.New()

	if !wantsThumb(httptest.NewRequest("GET", publicPhotoURL(profileID, photoID, true), nil)) {
		t.Error("wantsThumb no reconoce la miniatura de publicPhotoURL")
	}
	if !wantsThumb(httptest.NewRequest("GET", ownPhotoURL(photoID, true), nil)) {
		t.Error("wantsThumb no reconoce la miniatura de ownPhotoURL")
	}
	if wantsThumb(httptest.NewRequest("GET", publicPhotoURL(profileID, photoID, false), nil)) {
		t.Error("la foto completa no debe pedir miniatura")
	}
}

// La respuesta JSON no debe cambiar respecto a cuando las URLs se construían a mano.
func TestPhotoResponses_KeepURLs(t *testing.T) {
	profileID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ph := &Photo{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Position: 2, CreatedAt: time.Unix(0, 0).UTC()}

	self := toPhotoResponseSelf(ph)
	if self.URL != "/api/v1/profiles/me/photos/22222222-2222-2222-2222-222222222222/file" ||
		self.ThumbURL != self.URL+"?size=thumb" {
		t.Errorf("toPhotoResponseSelf: %+v", self)
	}
	pub := toPhotoResponsePublic(ph, profileID)
	if pub.URL != "/api/v1/profiles/11111111-1111-1111-1111-111111111111/photos/22222222-2222-2222-2222-222222222222/file" ||
		pub.ThumbURL != pub.URL+"?size=thumb" || pub.Position != 2 {
		t.Errorf("toPhotoResponsePublic: %+v", pub)
	}
}

func TestCompletionPercent(t *testing.T) {
	cases := []struct {
		optional int
		photo    bool
		want     int
	}{
		{0, false, 33}, // solo tener perfil: 4/12
		{0, true, 50},  // + foto: 6/12
		{3, false, 58}, // 7/12 = 58,3
		{6, true, 100}, // todo: 12/12
		{6, false, 83}, // 10/12 = 83,3
	}
	for _, tc := range cases {
		if got := completionPercent(tc.optional, tc.photo); got != tc.want {
			t.Errorf("completionPercent(%d, %v) = %d, se esperaba %d", tc.optional, tc.photo, got, tc.want)
		}
	}
}
