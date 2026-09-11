package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage guarda los ficheros en un directorio del propio
// contenedor (normalmente un volumen Docker, ver STORAGE_LOCAL_PATH).
// Es el driver por defecto (STORAGE_DRIVER=local), pensado solo para
// desarrollo: no sobrevive a un despliegue con múltiples instancias.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage crea el driver y asegura que el directorio base existe.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("storage: no se pudo crear %q: %w", basePath, err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

var _ Storage = (*LocalStorage)(nil)

func (s *LocalStorage) Save(ctx context.Context, key string, r io.Reader) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("storage: no se pudo crear el directorio: %w", err)
	}

	// Escritura a un fichero temporal + rename atómico: evita dejar un
	// fichero a medias si el proceso se interrumpe a mitad de escritura.
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("storage: no se pudo crear el fichero temporal: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("storage: error escribiendo: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("storage: error cerrando el fichero: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("storage: no se pudo mover el fichero final: %w", err)
	}

	return nil
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: no se pudo abrir el fichero: %w", err)
	}

	return f, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: no se pudo borrar el fichero: %w", err)
	}

	return nil
}

// resolve calcula la ruta absoluta de una key y comprueba que quede
// dentro de basePath, para blindarnos frente a keys con "../" aunque en
// la práctica las genera siempre el propio backend, nunca el usuario.
func (s *LocalStorage) resolve(key string) (string, error) {
	cleanKey := filepath.Clean("/" + key) // fuerza una ruta "absoluta" antes de limpiar ".."
	path := filepath.Join(s.basePath, cleanKey)

	if !strings.HasPrefix(path, filepath.Clean(s.basePath)+string(os.PathSeparator)) {
		return "", fmt.Errorf("storage: key inválida: %q", key)
	}

	return path, nil
}
