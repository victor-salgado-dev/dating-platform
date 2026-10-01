package profiles

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
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

// processedPhoto son los dos ficheros que se guardan por cada foto subida.
type processedPhoto struct {
	Full  []byte
	Thumb []byte
}

// processPhoto valida, normaliza y reduce una foto subida:
//   - decodifica de verdad (no se fía de cabeceras ni de la extensión),
//   - aplica la orientación EXIF y elimina TODOS los metadatos (GPS incluido)
//     al recodificar solo los píxeles,
//   - aplana la transparencia sobre blanco,
//   - reduce a photoMaxDimension (sin ampliar nunca) y genera la miniatura.
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
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, invalidField("photo", "No se pudo leer la imagen. Prueba con otra foto (JPG, PNG o WebP).")
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
	img = flattenOnWhite(img)

	// Fit nunca amplía: si la imagen ya cabe, devuelve una copia tal cual.
	full := imaging.Fit(img, photoMaxDimension, photoMaxDimension, imaging.Lanczos)
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

// flattenOnWhite sustituye la transparencia por fondo blanco: JPEG no tiene
// canal alfa y, sin esto, los píxeles transparentes saldrían negros.
func flattenOnWhite(img image.Image) image.Image {
	if o, ok := img.(interface{ Opaque() bool }); ok && o.Opaque() {
		return img
	}
	bg := imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)
	return imaging.Overlay(bg, img, image.Pt(0, 0), 1.0)
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("profiles: codificar JPEG: %w", err)
	}
	return buf.Bytes(), nil
}
