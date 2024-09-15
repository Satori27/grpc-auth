package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/Satori27/sso/internal/config"
	"github.com/Satori27/sso/internal/domain/models"
	"github.com/Satori27/sso/internal/storage"
	"github.com/lib/pq"
)

type Storage struct{
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error){
	dbInfo := cfg.DB
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
	"password=%s dbname=%s sslmode=disable", dbInfo.Host, dbInfo.Port, dbInfo.User, dbInfo.Password, dbInfo.Name)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return &Storage{db: db}, nil
}


func (s *Storage) SaveUser(ctx context.Context, email string, passhash []byte) (int64, error){
	const op = "storage.postgres.SaveUser"

	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id")
	if err!=nil{
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	
	res := stmt.QueryRowContext(ctx, email, string(passhash))
	var userID int64
	err = res.Scan(&userID)
	if err!=nil{
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pq.ErrorCode("23505") {
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}



func (s *Storage) User(ctx context.Context, email string) (models.User, error){
	const op = "storage.postgres.User"

	stmt, err := s.db.PrepareContext(ctx, "SELECT * FROM users WHERE email=$1")

	if err!=nil{
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}


	defer stmt.Close()

	row:=stmt.QueryRowContext(ctx, email)

	user:=models.User{}

	err = row.Scan(&user.ID, &user.Email, &user.PassHash)

	if err!=nil{
		if errors.Is(err, sql.ErrNoRows){
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}


	return user, nil
}

func (s *Storage) Role(ctx context.Context, userID int64) (string, error){
	const op = "storage.postgres.Role"

	stmt, err:=s.db.PrepareContext(ctx, "SELECT user_role FROM roles WHERE user_id=$1")

	if err!=nil{
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, userID)
	var role string
	err = row.Scan(&role)

	if err!=nil{
		if errors.Is(err, sql.ErrNoRows){
			return "", storage.ErrRoleIsEmpty
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}


func (s *Storage) App(ctx context.Context, appID int) (models.App, error){
	const op = "storage.postgres.App"

	stmt, err:=s.db.PrepareContext(ctx, "SELECT * FROM apps WHERE id=$1")

	if err!=nil{
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, appID)
	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)

	if err!=nil{
		if errors.Is(err, sql.ErrNoRows){
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}


func (s *Storage) SaveApp(ctx context.Context, name string, secret string, userID int64) (int64, error){
	const op = "storage.postgres.SaveApp"

	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO apps (name, secret) VALUES($1, $2) RETURNING id")

	if err!=nil{
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	row:=stmt.QueryRowContext(ctx, name, secret)
	var appID int64
	err = row.Scan(&appID)

	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code==pq.ErrorCode("23505"){
		return 0, storage.ErrAppNameExists
	}

	if err!=nil{
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return appID, nil
}