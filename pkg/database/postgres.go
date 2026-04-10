// pkg/database/postgres.go
package database

import (
	"fmt"
	"goScadaSur/pkg/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

// PostgresClient maneja la conexión con PostgreSQL
type PostgresClient struct {
	DB *gorm.DB
}

// NewPostgresClient inicializa y migra la base de datos
func NewPostgresClient(cfg *config.AppConfig) (*PostgresClient, error) {
	// Verificar si la configuración es válida/existe
	if cfg.Postgres.Host == "" || cfg.Postgres.Port == 0 {
		log.Println("[INFO] PostgreSQL no configurado. Modo de persistencia desactivado.")
		return nil, nil
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password, 
		cfg.Postgres.DBName, cfg.Postgres.Port, cfg.Postgres.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error conectando a postgres: %w", err)
	}

	// Migración automática de esquemas
	log.Println("[INFO] PostgreSQL conectado. Ejecutando migraciones...")
	err = db.AutoMigrate(&User{}, &Job{}, &Comment{}, &JobHistory{})
	if err != nil {
		return nil, fmt.Errorf("error migrando esquemas: %w", err)
	}

	// Sembrar usuario administrador por defecto si la tabla está vacía
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		log.Println("[INFO] No se encontraron usuarios. Creando administrador predeterminado...")
		hashedPassword, _ := HashPassword("admin123")
		adminUser := User{
			Username: "admin",
			Password: hashedPassword,
			Email:    "admin@example.com",
			FullName: "System Administrator",
			Role:     "admin",
		}
		db.Create(&adminUser)
		log.Println("[OK] Usuario 'admin' con clave 'admin123' creado correctamente.")
	}

	return &PostgresClient{DB: db}, nil
}

// HashPassword genera un hash seguro de la contraseña
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compara una contraseña con su hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
