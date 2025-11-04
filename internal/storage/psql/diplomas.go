package psql

import (
	"context"
	"gosmol/internal/domain"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DiplomasRepo struct {
	db *pgxpool.Pool
}

func NewDiplomasRepo(db *pgxpool.Pool) *DiplomasRepo {
	return &DiplomasRepo{db: db}
}

func (d *DiplomasRepo) SelectAllResource(limits int64, page int64) ([]domain.Diploma, error) {
	offset := (page - 1) * limits
	rows, err := d.db.Query(context.Background(), "SELECT id, name, description FROM diplomas LIMIT $1 OFFSET $2", limits, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diplomas []domain.Diploma
	for rows.Next() {
		var d domain.Diploma
		if err := rows.Scan(&d.ID, &d.Title, &d.Description); err != nil {
			return nil, err
		}

		diplomas = append(diplomas, d)
	}	

	return diplomas, nil
}

func (d *DiplomasRepo) SelectResource(id int64) (domain.Diploma, error) {
	var dip domain.Diploma
	err := d.db.QueryRow(context.Background(), 
		"SELECT id, title, description FROM diplomas WHERE id = $1", id).
		Scan(&dip.ID, &dip.Title, &dip.Description)

	return dip, err
}

func (d *DiplomasRepo) InsertResource(diploma domain.Diploma) (int64, error) {
	var id int64
	q := `
		INSERT INTO diplomas 
			(title, description) 
		VALUES 
			($1, $2) 
		RETURNING id
	`
	err := d.db.QueryRow(context.Background(), q, diploma.Title, diploma.Description).Scan(&id)
	// if err := d.db.QueryRow(context.Background(), q, diploma.Title, diploma.Description).Scan(&diploma.ID); err != nil {
	//   if pgErr, ok := err.(*pgconn.PgError); ok {
	//     newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
	//     fmt.Println(newErr)
	//     return "", nil
	//   }
	//   return "", err
	// }
	//
	return id, err
}

func (d *DiplomasRepo) RenovationResource(id int64, diploma domain.Diploma) (domain.Diploma, error) {
	_, err := d.db.Exec(context.Background(),
		"UPDATE diplomas SET name = $1, description = $2 WHERE id = $3",
		diploma.Title, diploma.Description, id)

	if err != nil {
		return diploma, err
	}

	return diploma, nil
}

func (d *DiplomasRepo) DestroyResource(id int64) error {
	_, err := d.db.Exec(context.Background(),
    "DELETE FROM diplomas WHERE id = $1", id)
	
	if err != nil {
		return err
	}

	return nil
}