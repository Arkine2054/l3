package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Image struct {
	ID            int64   `json:"id"`
	OrigFilename  string  `json:"orig_filename"`
	StoredPath    *string `json:"stored_path"`
	ProcessedPath *string `json:"processed_path"`
	ThumbPath     *string `json:"thumb_path"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

type ImagesRepo struct {
	DB *sql.DB
}

func NewImagesRepo() (*ImagesRepo, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" {
		host = "postgres"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres2"
	}
	if pass == "" {
		pass = "postgres2"
	}
	if name == "" {
		name = "images"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping error: %w", err)
	}

	repo := &ImagesRepo{DB: db}
	if err := repo.migrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Connected to DB and migrations applied successfully.")
	return repo, nil
}

func (r *ImagesRepo) migrate(db *sql.DB) error {
	dir := "./internal/migrations"
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	log.Println("Running goose migrations from:", dir)
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")
	return nil
}

func (r *ImagesRepo) Create(storedPath string, filename string, format string) (int64, error) {
	query := `
		INSERT INTO images (orig_filename, stored_path, status, format, created_at)
		VALUES ($1, $2, 'pending',$3, NOW())
		RETURNING id
	`
	var id int64
	err := r.DB.QueryRow(query, filename, storedPath, format).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ImagesRepo) UpdatePathsAndStatus(id int64, processedPath, thumbPath *string, status string) error {
	_, err := r.DB.Exec(`
		UPDATE images 
		SET processed_path = COALESCE($1, processed_path),
		    thumb_path = COALESCE($2, thumb_path),
		    status = $3,
		    updated_at=NOW()
		WHERE id = $4
	`, processedPath, thumbPath, status, id)
	return err
}

func (r *ImagesRepo) GetByID(id int64) (*Image, error) {
	row := r.DB.QueryRow(`
		SELECT id, stored_path, processed_path, thumb_path, status, created_at
		FROM images WHERE id = $1
	`, id)
	var img Image
	err := row.Scan(&img.ID, &img.StoredPath, &img.ProcessedPath, &img.ThumbPath, &img.Status, &img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *ImagesRepo) List() ([]Image, error) {
	rows, err := r.DB.Query(`
	SELECT id, orig_filename, stored_path, processed_path, thumb_path, status, created_at
	FROM images ORDER BY id DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		err := rows.Scan(
			&img.ID,
			&img.OrigFilename,
			&img.StoredPath,
			&img.ProcessedPath,
			&img.ThumbPath,
			&img.Status,
			&img.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *ImagesRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM images WHERE id = $1`, id)
	return err
}
