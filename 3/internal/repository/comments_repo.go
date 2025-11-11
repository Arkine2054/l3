package repository

import (
	"database/sql"

	"gitlab.com/arkine/l3/3/internal/models"
)

type CommentRepo struct {
	DB *sql.DB
}

// Создание комментария
func (r *CommentRepo) Create(c *models.Comment) error {
	query := `
		INSERT INTO comments (parent_id, author, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query, c.ParentID, c.Author, c.Content).
		Scan(&c.ID, &c.CreatedAt)
}

// Удаление комментария и всех дочерних
func (r *CommentRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`
		WITH RECURSIVE sub AS (
			SELECT id FROM comments WHERE id = $1
			UNION ALL
			SELECT c.id FROM comments c
			INNER JOIN sub s ON c.parent_id = s.id
		)
		DELETE FROM comments WHERE id IN (SELECT id FROM sub)
	`, id)
	return err
}

// Получение всех комментариев с вложенностью
func (r *CommentRepo) List() ([]*models.Comment, error) {
	rows, err := r.DB.Query(`
		SELECT id, parent_id, author, content, created_at
		FROM comments
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	all := []*models.Comment{}
	lookup := make(map[int64]*models.Comment)

	for rows.Next() {
		var c models.Comment
		var parentID sql.NullInt64
		var author sql.NullString

		if err := rows.Scan(&c.ID, &parentID, &author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.Int64
		}
		if author.Valid {
			c.Author = &author.String
		}

		c.Children = []*models.Comment{}
		all = append(all, &c)
		lookup[c.ID] = &c
	}

	// Собираем дерево
	var roots []*models.Comment
	for _, c := range all {
		if c.ParentID != nil {
			if parent, ok := lookup[*c.ParentID]; ok {
				parent.Children = append(parent.Children, c)
			}
		} else {
			roots = append(roots, c)
		}
	}

	return roots, nil
}

// Поиск по содержимому
func (r *CommentRepo) Search(term string, limit, offset int) ([]*models.Comment, error) {
	rows, err := r.DB.Query(`
		SELECT id, parent_id, author, content, created_at
		FROM comments
		WHERE content ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, term, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*models.Comment{}
	for rows.Next() {
		var c models.Comment
		var parentID sql.NullInt64
		var author sql.NullString

		if err := rows.Scan(&c.ID, &parentID, &author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.Int64
		}
		if author.Valid {
			c.Author = &author.String
		}

		results = append(results, &c)
	}

	return results, nil
}
