package repository

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"gitlab.com/arkine/l3/7/internal/models"
	"gitlab.com/arkine/l3/7/internal/utils"
)

var query string

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	return &Repo{DB: db}
}

func (r *Repo) SetCurrentUser(ctx context.Context, userID int64) error {
	_, err := r.DB.ExecContext(ctx,
		"SELECT set_config('app.current_user_id', $1, true)",
		strconv.FormatInt(userID, 10),
	)
	return err
}

func (r *Repo) setCurrentUserTx(ctx context.Context, tx *sql.Tx) error {
	claims := utils.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil
	}

	uid := utils.GetUserIDFromClaims(claims)
	if uid <= 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", strconv.FormatInt(uid, 10))
	if err != nil {
		log.Printf("[Repo] set_config err: %v", err)
	}
	return err
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	query = `SELECT id, username, password, role FROM users WHERE username=$1`
	err := r.DB.QueryRowContext(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Role)
	if err != nil {
		log.Printf("[GetUserByUsername] QueryRow err %v", err)
		return nil, err
	}
	return &u, nil
}

func (r *Repo) CreateUser(ctx context.Context, user models.User) error {
	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	query = `INSERT INTO users(username, password, role) VALUES($1, $2, $3)`
	_, err = r.DB.ExecContext(ctx, query, user.Username, hashed, user.Role)
	if err != nil {
		log.Printf("[CreateUser] Exec err %v", err)
	}
	return err
}

func (r *Repo) GetAllItems(ctx context.Context) ([]models.Item, error) {
	query = `SELECT id, name, quantity, price FROM items ORDER BY id`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[GetAllItems] Query err %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var i models.Item
		if err = rows.Scan(&i.ID, &i.Name, &i.Quantity, &i.Price); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *Repo) CreateItem(ctx context.Context, i models.Item) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := r.setCurrentUserTx(ctx, tx); err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[CreateItem]CurrentUserTX rollback err %v", err)
		}
		return err
	}

	query = `INSERT INTO items(name, quantity, price) VALUES($1,$2,$3)`
	_, err = tx.ExecContext(ctx, query, i.Name, i.Quantity, i.Price)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[CreateItem] Rollback err %v", err)
		}
		log.Printf("[CreateItem] Exec err %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateItem(ctx context.Context, id int64, i models.Item) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := r.setCurrentUserTx(ctx, tx); err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[UpdateItem] CurrentUser Rollback err %v", err)
		}
		return err
	}

	query = `UPDATE items SET name=$1, quantity=$2, price=$3 WHERE id=$4`
	_, err = tx.ExecContext(ctx, query, i.Name, i.Quantity, i.Price, id)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[UpdateItem] Rollback err %v", err)
		}
		log.Printf("[UpdateItem] Exec err %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repo) DeleteItem(ctx context.Context, id int64) error {

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := r.setCurrentUserTx(ctx, tx); err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[DeleteItem] currentUser Rollback err %v", err)
		}
		return err
	}

	query = `DELETE FROM items WHERE id=$1`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Printf("[DeleteItem] Rollback err %v", err)
		}
		log.Printf("[DeleteItem] Exec err %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetHistoryForItem(ctx context.Context, id int64) ([]models.History, error) {
	query = `SELECT id, item_id, user_id, action, old_data, new_data, timestamp FROM history WHERE item_id=$1 ORDER BY timestamp DESC`
	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		log.Printf("[GetHistoryForItem] Query err %v", err)
		return nil, err
	}
	defer rows.Close()

	var history []models.History
	for rows.Next() {
		var h models.History
		if err := rows.Scan(&h.ID, &h.ItemID, &h.UserID, &h.Action, &h.OldData, &h.NewData, &h.Timestamp); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}
