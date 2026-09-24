// Package integration reúne las pruebas de integración que cruzan varios
// módulos del backend y necesitan una base de datos PostgreSQL real.
//
// Viven en un paquete propio (en lugar de dentro de favorites, likes,
// visits o blocking) por dos motivos: poder importar los cuatro
// repositorios sin crear ciclos de importación, y compartir un único
// conjunto de helpers de conexión y limpieza.
//
// Las pruebas concretas están en los ficheros *_integration_test.go, que
// llevan el build tag `integration` y se ejecutan con:
//
//	go test -tags=integration ./internal/integration/...
//
// Requieren la variable de entorno TEST_DATABASE_URL; si no está definida,
// los tests se saltan (t.Skip) en lugar de fallar.
package integration
