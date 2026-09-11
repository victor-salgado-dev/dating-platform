package storage

import "fmt"

// New construye el Storage configurado mediante STORAGE_DRIVER. En V1
// solo existe "local". Una fase de producción posterior podrá añadir un
// driver S3-compatible aquí sin que profiles ni ningún otro consumidor
// de Storage cambien una sola línea.
func New(driver, localPath string) (Storage, error) {
	switch driver {
	case "", "local":
		return NewLocalStorage(localPath)
	default:
		return nil, fmt.Errorf("storage: driver desconocido %q", driver)
	}
}
