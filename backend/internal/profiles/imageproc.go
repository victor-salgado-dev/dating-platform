package profiles

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // registra el decodificador PNG para image.Decode
	"io"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // registra el decodificador WebP (solo lectura)
)

// Parámetros del procesado de fotos. Se ajustan aquí, en un único sitio.
const (
	// Lado largo máximo de la foto que se guarda y de la miniatura.
	photoMaxDimension = 1600
	thumbMaxDimension = 480

	// Calidad JPEG (0-100). 82 es visualmente indistinguible del original
	// en pantalla y reduce mucho el peso; la miniatura admite algo menos.
	photoJPEGQuality = 82
	thumbJPEGQuality = 78

	// Tope de píxeles de la imagen de ENTRADA. Protege la memoria: un PNG
	// de 100 MB de píxeles ocupa muy poco en disco pero ~400 MB al decodificar.
	photoMaxPixels = 40_000_000

	// Cuántas fotos se decodifican a la vez. El resto espera su turno.
	maxConcurrentPhotoProcessing = 2

	// Lo que producimos siempre, sea cual sea el formato de entrada.
	processedContentType = "image/jpeg"
	processedExt         = ".jpg"
	thumbKeySuffix       = "_thumb"
)

var photoProcessingSlots = make(chan struct{}, maxConcurrentPhotoProcessing)

// allowedPhotoFormats son los formatos de ENTRADA admitidos (los nombres que
// devuelve image.DecodeConfig). imaging registra además GIF, BMP y TIFF, así que
// la garantía de "solo JPEG, PNG o WebP" se aplica aquí, junto a la decodificación,
// y no solo en la detección previa del service.
var allowedPhotoFormats = map[string]bool{"jpeg": true, "png": true, "webp": true}

// processedPhoto son los dos ficheros que se guardan por cada foto subida.
type processedPhoto struct {
	Full  []byte
	Thumb []byte
}

// processPhoto valida, normaliza y reduce una foto subida:
//   - decodifica de verdad (no se fía de cabeceras ni de la extensión),
//   - aplica la orientación EXIF y elimina TODOS los metadatos (GPS incluido)
//     al recodificar solo los píxeles,
//   - reduce a photoMaxDimension (sin ampliar nunca) y genera la miniatura,
//   - aplana la transparencia sobre blanco.
//
// El orden importa para la memoria: se reduce ANTES de aplanar. Aplanar la
// imagen original obligaba a mantener tres copias a tamaño completo (original,
// fondo y resultado): una PNG transparente de 16 MP llegaba a 245 MB de pico.
// Reducida primero, el aplanado trabaja sobre ~1600 px y se hace en el sitio.
//
// Devuelve errores de validación (invalidField) para ficheros que no son una
// imagen utilizable, y errores normales para fallos internos.
func processPhoto(ctx context.Context, r io.Reader) (*processedPhoto, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxPhotoSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("profiles: leer la foto: %w", err)
	}
	if int64(len(data)) > MaxPhotoSizeBytes {
		return nil, invalidField("photo", "La foto es demasiado grande.")
	}

	// Lee solo las dimensiones, sin decodificar los píxeles: así se rechazan
	// las "bombas de descompresión" antes de gastar memoria.
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, invalidField("photo", "No se pudo leer la imagen. Prueba con otra foto (JPG, PNG o WebP).")
	}
	if !allowedPhotoFormats[format] {
		return nil, invalidField("photo", "Formato de imagen no admitido. Usa JPG, PNG o WebP.")
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > photoMaxPixels {
		return nil, invalidField("photo", "Las dimensiones de la imagen no son válidas o son demasiado grandes.")
	}

	select {
	case photoProcessingSlots <- struct{}{}:
		defer func() { <-photoProcessingSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, invalidField("photo", "No se pudo leer la imagen. Prueba con otra foto (JPG, PNG o WebP).")
	}

	// Fit nunca amplía: si la imagen ya cabe, devuelve una copia tal cual (por
	// eso full es siempre nuestro y puede modificarse en el sitio). El remuestreo
	// pondera por alfa, así que reducir antes de aplanar no deja halos oscuros.
	full := imaging.Fit(img, photoMaxDimension, photoMaxDimension, imaging.Lanczos)
	flattenOnWhite(full)
	thumb := imaging.Fit(full, thumbMaxDimension, thumbMaxDimension, imaging.Lanczos)

	fullBytes, err := encodeJPEG(full, photoJPEGQuality)
	if err != nil {
		return nil, err
	}
	thumbBytes, err := encodeJPEG(thumb, thumbJPEGQuality)
	if err != nil {
		return nil, err
	}
	return &processedPhoto{Full: fullBytes, Thumb: thumbBytes}, nil
}

// flattenOnWhite sustituye la transparencia por fondo blanco, EN EL SITIO:
// JPEG no tiene canal alfa y, sin esto, los píxeles transparentes saldrían
// negros. Respeta el stride, así que funciona también sobre una subimagen.
func flattenOnWhite(img *image.NRGBA) {
	b := img.Bounds()
	w := b.Dx()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := img.PixOffset(b.Min.X, y)
		row := img.Pix[i : i+4*w]
		for x := 0; x < len(row); x += 4 {
			a := uint32(row[x+3])
			if a == 255 {
				continue
			}
			inv := 255 - a
			row[x+0] = uint8((uint32(row[x+0])*a + 255*inv + 127) / 255)
			row[x+1] = uint8((uint32(row[x+1])*a + 255*inv + 127) / 255)
			row[x+2] = uint8((uint32(row[x+2])*a + 255*inv + 127) / 255)
			row[x+3] = 255
		}
	}
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("profiles: codificar JPEG: %w", err)
	}
	return buf.Bytes(), nil
}
