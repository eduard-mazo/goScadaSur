package database

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "admin123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hasheando contraseña: %v", err)
	}

	if hash == password {
		t.Errorf("El hash no debería ser igual a la contraseña plana")
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("La validación del hash falló")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("La validación debería fallar para una contraseña incorrecta")
	}
}
