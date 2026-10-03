package profiles

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"runtime"
	"testing"
)

// --- Generadores de imágenes de prueba ---------------------------------------

func solidJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 255 / w), uint8(y * 255 / h), 90, 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// halfOpaquePNG: mitad izquierda del color left (opaca) y derecha TRANSPARENTE con
// RGB negro (0,0,0,0), que es el caso que dejaría un halo oscuro si el remuestreo
// no ponderase por alfa.
func halfOpaquePNG(t *testing.T, w, h int, left color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w/2; x++ {
			img.SetNRGBA(x, y, left)
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// withEXIFOrientation inserta un segmento APP1/EXIF con la etiqueta Orientation.
func withEXIFOrientation(t *testing.T, jpg []byte, orientation uint16) []byte {
	t.Helper()
	tiff := []byte{'M', 'M', 0, 0x2a, 0, 0, 0, 8, // cabecera big-endian, IFD en offset 8
		0, 1, // una entrada
		0x01, 0x12, 0, 3, 0, 0, 0, 1, byte(orientation >> 8), byte(orientation), 0, 0, // Orientation, SHORT
		0, 0, 0, 0} // sin más IFDs
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}
	seg = append(seg, payload...)
	return append(append(append([]byte{}, jpg[:2]...), seg...), jpg[2:]...)
}

// pngHeaderOnly es un PNG que SOLO declara dimensiones enormes (IHDR + IEND):
// image.DecodeConfig las lee, pero decodificarlo exigiría gigas.
func pngHeaderOnly(w, h uint32) []byte {
	chunk := func(kind string, data []byte) []byte {
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, uint32(len(data)))
		b.WriteString(kind)
		b.Write(data)
		_ = binary.Write(&b, binary.BigEndian, crc32.ChecksumIEEE(append([]byte(kind), data...)))
		return b.Bytes()
	}
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], w)
	binary.BigEndian.PutUint32(ihdr[4:], h)
	ihdr[8], ihdr[9] = 8, 2 // 8 bits, RGB
	out := []byte("\x89PNG\r\n\x1a\n")
	out = append(out, chunk("IHDR", ihdr)...)
	return append(out, chunk("IEND", nil)...)
}

func decodeJPEG(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("la salida debe ser un JPEG válido: %v", err)
	}
	return img
}

func dims(img image.Image) (int, int) { b := img.Bounds(); return b.Dx(), b.Dy() }

// --- Tests -----------------------------------------------------------------------

func TestProcessPhoto_ResizesAndNeverUpscales(t *testing.T) {
	cases := []struct {
		name                         string
		w, h                         int
		wantW, wantH, thumbW, thumbH int
	}{
		{"grande se reduce", 3200, 2400, 1600, 1200, 480, 360},
		{"vertical se reduce", 2400, 3200, 1200, 1600, 360, 480},
		{"pequeña no se amplía", 800, 600, 800, 600, 480, 360},
		{"diminuta no se amplía", 100, 80, 100, 80, 100, 80},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := processPhoto(context.Background(), bytes.NewReader(solidJPEG(t, tc.w, tc.h)))
			if err != nil {
				t.Fatalf("processPhoto: %v", err)
			}
			if w, h := dims(decodeJPEG(t, out.Full)); w != tc.wantW || h != tc.wantH {
				t.Errorf("foto = %dx%d, se esperaba %dx%d", w, h, tc.wantW, tc.wantH)
			}
			if w, h := dims(decodeJPEG(t, out.Thumb)); w != tc.thumbW || h != tc.thumbH {
				t.Errorf("miniatura = %dx%d, se esperaba %dx%d", w, h, tc.thumbW, tc.thumbH)
			}
		})
	}
}

// Fondo transparente -> blanco, y sin halo oscuro en el borde. La imagen se
// REDUCE (3200 -> 1600), así que se ejercita el remuestreo sobre píxeles
// transparentes con RGB negro, justo lo que cambia al reducir antes de aplanar.
func TestProcessPhoto_FlattensTransparencyOnWhiteWithoutDarkHalo(t *testing.T) {
	data := halfOpaquePNG(t, 3200, 1600, color.NRGBA{255, 255, 255, 255})
	out, err := processPhoto(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("processPhoto: %v", err)
	}

	for name, raw := range map[string][]byte{"foto": out.Full, "miniatura": out.Thumb} {
		img := decodeJPEG(t, raw)
		b := img.Bounds()
		minLum := uint32(255)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r, g, bl, _ := img.At(x, y).RGBA()
				if lum := (r + g + bl) / 3 >> 8; lum < minLum {
					minLum = lum
				}
			}
		}
		// Blanco sobre transparente debe quedar blanco en TODA la imagen.
		if minLum < 240 {
			t.Errorf("%s: hay píxeles de luminancia %d (halo oscuro o fondo negro)", name, minLum)
		}
	}
}

func TestProcessPhoto_TransparentAreaBecomesWhiteAndOpaqueKeepsColor(t *testing.T) {
	data := halfOpaquePNG(t, 400, 200, color.NRGBA{200, 0, 0, 255}) // rojo opaco | transparente
	out, err := processPhoto(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("processPhoto: %v", err)
	}
	img := decodeJPEG(t, out.Full)

	r, g, b, _ := img.At(300, 100).RGBA() // zona transparente
	if r>>8 < 250 || g>>8 < 250 || b>>8 < 250 {
		t.Errorf("la zona transparente debe ser blanca, es (%d,%d,%d)", r>>8, g>>8, b>>8)
	}
	r, g, b, _ = img.At(50, 100).RGBA() // zona opaca roja
	if r>>8 < 170 || g>>8 > 50 || b>>8 > 50 {
		t.Errorf("la zona opaca debe seguir roja, es (%d,%d,%d)", r>>8, g>>8, b>>8)
	}
}

func TestProcessPhoto_AppliesEXIFOrientationAndStripsMetadata(t *testing.T) {
	landscape := solidJPEG(t, 300, 200)
	withExif := withEXIFOrientation(t, landscape, 6) // 6 = girar 90° en sentido horario
	if !bytes.Contains(withExif, []byte("Exif")) {
		t.Fatal("la entrada de prueba debería llevar EXIF")
	}

	out, err := processPhoto(context.Background(), bytes.NewReader(withExif))
	if err != nil {
		t.Fatalf("processPhoto: %v", err)
	}

	if w, h := dims(decodeJPEG(t, out.Full)); w != 200 || h != 300 {
		t.Errorf("con orientación 6 una foto 300x200 debe quedar 200x300 (vertical), es %dx%d", w, h)
	}
	if bytes.Contains(out.Full, []byte("Exif")) || bytes.Contains(out.Thumb, []byte("Exif")) {
		t.Error("la salida no debe conservar metadatos EXIF (GPS incluido)")
	}
}

func TestProcessPhoto_RejectsNonImagesAndDisallowedFormats(t *testing.T) {
	var gifBuf bytes.Buffer
	_ = gif.Encode(&gifBuf, image.NewPaletted(image.Rect(0, 0, 10, 10), color.Palette{color.Black, color.White}), nil)

	cases := map[string][]byte{
		"vacío":                                {},
		"texto":                                []byte("esto no es una imagen"),
		"PDF":                                  []byte("%PDF-1.7 ..."),
		"JPEG truncado":                        solidJPEG(t, 100, 100)[:50],
		"GIF (decodificable pero no admitido)": gifBuf.Bytes(),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := processPhoto(context.Background(), bytes.NewReader(data))
			if err == nil || out != nil {
				t.Fatalf("se esperaba un error de validación, se obtuvo out=%v err=%v", out != nil, err)
			}
		})
	}
}

// Una imagen que declara 100 MP se rechaza leyendo solo la cabecera, sin gastar memoria.
func TestProcessPhoto_RejectsHugeDimensionsWithoutDecoding(t *testing.T) {
	data := pngHeaderOnly(10000, 10000) // 100 MP > photoMaxPixels

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	out, err := processPhoto(context.Background(), bytes.NewReader(data))
	runtime.ReadMemStats(&after)

	if err == nil || out != nil {
		t.Fatal("se esperaba rechazar una imagen de 100 MP")
	}
	if mb := float64(after.TotalAlloc-before.TotalAlloc) / 1e6; mb > 20 {
		t.Errorf("el rechazo asignó %.0f MB: parece que se intentó decodificar", mb)
	}
}

func TestProcessPhoto_RespectsCancelledContextWhenBusy(t *testing.T) {
	for i := 0; i < maxConcurrentPhotoProcessing; i++ {
		photoProcessingSlots <- struct{}{}
	}
	defer func() {
		for i := 0; i < maxConcurrentPhotoProcessing; i++ {
			<-photoProcessingSlots
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := processPhoto(ctx, bytes.NewReader(solidJPEG(t, 100, 100))); err == nil {
		t.Fatal("con todos los slots ocupados y el contexto cancelado debe devolver error")
	}
}

// Regresión de memoria: una PNG transparente de 16 MP (64 MB decodificada) llegaba
// a asignar ~245 MB cuando se aplanaba antes de reducir. Ahora ~117 MB.
func TestProcessPhoto_MemoryOnLargeTransparentPNG(t *testing.T) {
	if testing.Short() {
		t.Skip("genera una imagen de 16 MP")
	}
	data := halfOpaquePNG(t, 4000, 4000, color.NRGBA{30, 120, 200, 255})

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, err := processPhoto(context.Background(), bytes.NewReader(data)); err != nil {
		t.Fatalf("processPhoto: %v", err)
	}
	runtime.ReadMemStats(&after)

	const raw = 4000 * 4000 * 4 // bytes de la imagen decodificada
	allocated := after.TotalAlloc - before.TotalAlloc
	if limit := uint64(2.7 * raw); allocated > limit {
		t.Errorf("asignó %d MB (límite %d MB, ~%.1f× la imagen): vuelve a haber copias a tamaño completo",
			allocated>>20, limit>>20, float64(allocated)/raw)
	}
}

func TestFlattenOnWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	img.SetNRGBA(0, 0, color.NRGBA{10, 20, 30, 255}) // opaco: intacto
	img.SetNRGBA(1, 0, color.NRGBA{0, 0, 0, 0})      // transparente -> blanco
	img.SetNRGBA(2, 0, color.NRGBA{0, 0, 0, 128})    // negro al 50% -> gris medio
	img.SetNRGBA(3, 0, color.NRGBA{255, 0, 0, 0})    // transparente con color -> blanco

	flattenOnWhite(img)

	want := []color.NRGBA{{10, 20, 30, 255}, {255, 255, 255, 255}, {127, 127, 127, 255}, {255, 255, 255, 255}}
	for x, w := range want {
		if got := img.NRGBAAt(x, 0); got != w {
			t.Errorf("píxel %d = %v, se esperaba %v", x, got, w)
		}
	}
}

// Sobre una subimagen (stride mayor que el ancho) solo se toca su rectángulo.
func TestFlattenOnWhite_RespectsStrideOfSubImage(t *testing.T) {
	full := image.NewNRGBA(image.Rect(0, 0, 6, 3)) // todo transparente (0,0,0,0)
	sub := full.SubImage(image.Rect(2, 1, 4, 2)).(*image.NRGBA)

	flattenOnWhite(sub)

	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			inside := x >= 2 && x < 4 && y == 1
			got := full.NRGBAAt(x, y)
			if inside && got != (color.NRGBA{255, 255, 255, 255}) {
				t.Errorf("(%d,%d) dentro de la subimagen debe ser blanco, es %v", x, y, got)
			}
			if !inside && got != (color.NRGBA{}) {
				t.Errorf("(%d,%d) fuera de la subimagen no debe tocarse, es %v", x, y, got)
			}
		}
	}
}
