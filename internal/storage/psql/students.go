package psql

import (
	"context"
	"errors"
	"gosmol/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentsRepo struct {
	db *pgxpool.Pool
}

func NewStudentsRepo(db *pgxpool.Pool) *StudentsRepo {
	return &StudentsRepo{db: db}
}

func (s *StudentsRepo) InsertStudents(students domain.Student) (int64, error) {
	var id int64
	err := s.db.QueryRow(context.Background(), 
		"INSERT INTO users (firstname, lastname, email, password_hash) VALUES ($1, $2) RETURNING id",
		students.Firstname, students.Lastname, students.Email, students.PasswordHash).
		Scan(&id)
	
	return id, err
}

func (s *StudentsRepo) SelectStudents(email string) (domain.Student, error) {
	var stud domain.Student
	err := s.db.QueryRow(context.Background(), 
		"SELECT id, firstname, lastname, email, password_hash, created_at FROM users WHERE email = $1, email").
		Scan(&stud.ID, &stud.Firstname, &stud.Lastname, &stud.Email, &stud.PasswordHash, &stud.CreatedAt)
	
	return stud, err
}

func (s *StudentsRepo) RefreshStore(userID int64, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(context.Background(), 
		"INSERT INTO refresh_token (user_id, token, expires_at) VALUES ($1, $2, $3)", 
		userID, token, expiresAt)
	
	return err
}

func (s *StudentsRepo) RefreshGet(token string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(context.Background(), 
		"SELECT user_id, expires_at FROM refresh_token WHERE token = $1", token).
		Scan(&userID, &expiresAt)

	if err != nil {
		return 0, err
	}

	if time.Now().After(expiresAt) {
		s.RefreshDelete(token)
		return 0, errors.New("token expired")
	}

	return userID, nil
}

func (s *StudentsRepo) RefreshDelete(token string) error {
	_, err := s.db.Exec(context.Background(), 
		"DELETE FROM refresh_tokens WHERE token = $1", token)

	return err
}