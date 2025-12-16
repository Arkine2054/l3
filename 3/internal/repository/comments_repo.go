package repository

import (
	"database/sql"
	"log"

	"gitlab.com/arkine/l3/3/internal/models"
)

type CommentRepo struct {
	DB *sql.DB
}

func (r *CommentRepo) Create(c *models.Comment) error {
	query := `
		INSERT INTO comments (parent_id, author, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query, c.ParentID, c.Author, c.Content).
		Scan(&c.ID, &c.CreatedAt)
}

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

func (r *CommentRepo) List() ([]*models.Comment, error) {
	rows, err := r.DB.Query(`
		SELECT id, parent_id, author, content, created_at
		FROM comments
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing list rows: %v", err)
		}
	}(rows)

	commentsMap := make(map[int64]*models.Comment)
	var roots []*models.Comment

	for rows.Next() {
		var c models.Comment
		var parentID sql.NullInt64
		if err := rows.Scan(&c.ID, &parentID, &c.Author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.Int64
		}

		c.Children = []*models.Comment{}
		commentsMap[c.ID] = &c

		if c.ParentID == nil {
			roots = append(roots, &c)
		} else {
			parent, ok := commentsMap[*c.ParentID]
			if ok {
				parent.Children = append(parent.Children, &c)
			} else {
				roots = append(roots, &c)
			}
		}
	}

	return roots, nil
}

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
