package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config son los parámetros necesarios para hablar con cualquier
// object storage compatible con la API de S3 (AWS S3, MinIO,
// DigitalOcean Spaces, Backblaze B2...).
type S3Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	ForcePathStyle bool
}

// S3Storage implementa Storage sobre un bucket S3-compatible. Es el
// driver pensado para producción (STORAGE_DRIVER=s3): a diferencia de
// LocalStorage, sobrevive a tener varias instancias del backend y no
// depende del disco de un contenedor concreto.
type S3Storage struct {
	client *minio.Client
	bucket string
}

// NewS3Storage crea el driver y comprueba que el bucket existe y es
// accesible, para fallar pronto (al arrancar) en vez de en el primer
// intento de subida de un usuario real.
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("storage: configuración S3 incompleta (endpoint/bucket/access key/secret key)")
	}

	lookup := minio.BucketLookupAuto
	if cfg.ForcePathStyle {
		lookup = minio.BucketLookupPath
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: lookup,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: no se pudo crear el cliente S3: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: no se pudo comprobar el bucket %q: %w", cfg.Bucket, err)
	}
	if !exists {
		return nil, fmt.Errorf("storage: el bucket %q no existe (créalo antes de arrancar en producción)", cfg.Bucket)
	}

	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

var _ Storage = (*S3Storage)(nil)

func (s *S3Storage) Save(ctx context.Context, key string, r io.Reader) error {
	// Tamaño -1: tamaño desconocido, minio-go se encarga de subirlo en
	// streaming (multipart si hiciera falta) sin que tengamos que leer
	// todo el fichero en memoria antes.
	_, err := s.client.PutObject(ctx, s.bucket, key, r, -1, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage: subir a S3: %w", err)
	}
	return nil
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage: abrir de S3: %w", err)
	}

	// GetObject es perezoso: la petición HTTP real no se dispara hasta
	// el primer Read/Stat. Llamamos a Stat aquí para poder devolver
	// ErrNotFound de forma consistente con LocalStorage, en vez de que
	// el error aparezca más tarde, en un Read que el llamador no espera
	// que pueda fallar por "no existe".
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		resp := minio.ToErrorResponse(err)
		if resp.Code == "NoSuchKey" {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: comprobar objeto en S3: %w", err)
	}

	return obj, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	// RemoveObject de S3 ya es idempotente (borrar algo que no existe
	// no es un error), igual que exige la interfaz Storage.
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("storage: borrar de S3: %w", err)
	}
	return nil
}
