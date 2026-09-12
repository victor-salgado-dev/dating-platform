// Comando promote-admin promociona una cuenta ya registrada al rol
// 'admin'. Es la única forma soportada de crear un administrador: no
// hay registro público de admins, y así evitamos tener que entrar a
// mano en psql (sección 6 del documento de proyecto).
//
// Uso: go run ./cmd/promote-admin --email=persona@example.com
// (normalmente vía `make admin-promote email=persona@example.com`)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/db"
)

func main() {
	email := flag.String("email", "", "Email de la cuenta a promocionar a admin")
	flag.Parse()

	if *email == "" {
		fmt.Fprintln(os.Stderr, "uso: promote-admin --email=persona@example.com")
		os.Exit(1)
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.Postgres)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo conectar a la base de datos: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	const query = `
		UPDATE users SET role = 'admin'
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`

	tag, err := pool.Exec(ctx, query, *email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo actualizar el rol: %v\n", err)
		os.Exit(1)
	}
	if tag.RowsAffected() == 0 {
		fmt.Fprintf(os.Stderr, "no se encontró ninguna cuenta activa con email %q\n", *email)
		os.Exit(1)
	}

	fmt.Printf("%s ahora es admin.\n", *email)
}
