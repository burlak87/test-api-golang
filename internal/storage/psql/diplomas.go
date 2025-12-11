package psql

import (
	"context"
	"fmt"
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
    fmt.Printf("DEBUG DIPLOMAS: Getting all diplomas, limit: %d, offset: %d\n", limits, offset)
    
    rows, err := d.db.Query(context.Background(), 
        "SELECT id, title, description FROM diplomas LIMIT $1 OFFSET $2", limits, offset)
    if err != nil {
        fmt.Printf("DEBUG DIPLOMAS: Error querying diplomas: %v\n", err)
        return nil, err
    }
    defer rows.Close()

    var diplomas []domain.Diploma
    for rows.Next() {
        var diploma domain.Diploma
        if err := rows.Scan(&diploma.ID, &diploma.Title, &diploma.Description); err != nil {
            fmt.Printf("DEBUG DIPLOMAS: Error scanning diploma: %v\n", err)
            return nil, err
        }
        fmt.Printf("DEBUG DIPLOMAS: Found diploma - ID: %d, Title: %s\n", diploma.ID, diploma.Title)
        diplomas = append(diplomas, diploma)
    }    

    fmt.Printf("DEBUG DIPLOMAS: Total diplomas found: %d\n", len(diplomas))
    return diplomas, nil
}

func (d *DiplomasRepo) InsertResource(diploma domain.Diploma) (int64, error) {
    var id int64
    fmt.Printf("DEBUG DIPLOMA INSERT: Starting - Title: %s\n", diploma.Title)
    
    query := `INSERT INTO diplomas (title, description) VALUES ($1, $2) RETURNING id`
    
    err := d.db.QueryRow(context.Background(), query, diploma.Title, diploma.Description).Scan(&id)
    if err != nil {
        fmt.Printf("DEBUG DIPLOMA INSERT: ERROR: %v\n", err)
        return 0, err
    }
    
    fmt.Printf("DEBUG DIPLOMA INSERT: SUCCESS - Inserted with ID: %d\n", id)
    return id, nil
}

func (d *DiplomasRepo) SelectResource(id int64) (domain.Diploma, error) {
    var diploma domain.Diploma
    fmt.Printf("DEBUG DIPLOMA SELECT: Getting diploma by ID: %d\n", id)
    
    err := d.db.QueryRow(context.Background(), 
        "SELECT id, title, description FROM diplomas WHERE id = $1", id).
        Scan(&diploma.ID, &diploma.Title, &diploma.Description)

    if err != nil {
        fmt.Printf("DEBUG DIPLOMA SELECT: ERROR: %v\n", err)
        return domain.Diploma{}, err
    }
    
    fmt.Printf("DEBUG DIPLOMA SELECT: SUCCESS - Diploma ID: %d, Title: %s\n", diploma.ID, diploma.Title)
    return diploma, nil
}

func (d *DiplomasRepo) RenovationResource(id int64, diploma domain.Diploma) (domain.Diploma, error) {
    fmt.Printf("DEBUG STORAGE DIPLOMA UPDATE: Updating diploma ID: %d\n", id)
    
    _, err := d.db.Exec(context.Background(),
        "UPDATE diplomas SET title = $1, description = $2 WHERE id = $3",
        diploma.Title, diploma.Description, id)

    if err != nil {
        fmt.Printf("DEBUG STORAGE DIPLOMA UPDATE: ERROR: %v\n", err)
        return domain.Diploma{}, err
    }
    
    updatedDiploma := domain.Diploma{
        ID:          id,
        Title:       diploma.Title,
        Description: diploma.Description,
    }
    
    fmt.Printf("DEBUG STORAGE DIPLOMA UPDATE: SUCCESS - Updated diploma ID: %d\n", id)
    return updatedDiploma, nil
}

func (d *DiplomasRepo) DestroyResource(id int64) error {
    fmt.Printf("DEBUG DIPLOMAS: Deleting diploma ID: %d\n", id)
    
    _, err := d.db.Exec(context.Background(), "DELETE FROM diplomas WHERE id = $1", id)   
    if err != nil {
        fmt.Printf("DEBUG DIPLOMAS: Error deleting diploma: %v\n", err)
        return err
    }
    
    fmt.Printf("DEBUG DIPLOMAS: Diploma deleted successfully\n")
    return nil
}