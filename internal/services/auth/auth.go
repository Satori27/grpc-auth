package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Satori27/sso/internal/domain/models"
	"github.com/Satori27/sso/internal/lib/jwt"
	"github.com/Satori27/sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const (
	ADMIN = "admin"
)

type Auth struct{
	log *slog.Logger
	userSaver UserSaver
	userProvider UserProvider
	appProvider AppProvider
	appSaver AppSaver
	tokenTTL time.Duration
}

var (
	ErrNoRightsForOperation = errors.New("you don't have rights for this operation")
	ErrInvalidCredentials = errors.New("invalid credentials")
)


type UserSaver interface{
	SaveUser(ctx context.Context, email string, passHash []byte)(uid int64, err error)
} 

type UserProvider interface{
	User(ctx context.Context, email string) (models.User, error)
	Role(ctx context.Context, userID int64) (string, error)
}


type AppProvider interface{
	App(ctx context.Context, appID int) (models.App, error)
}

type AppSaver interface{
	SaveApp(ctx context.Context, name string, secret string, userID int64) (int64, error)
}

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, appSaver AppSaver, tokenTTL time.Duration) *Auth{
	return &Auth{
		userSaver:userSaver,
		userProvider: userProvider,
		appProvider: appProvider,
		appSaver: appSaver,
		log: log,
		tokenTTL: tokenTTL,
	}
}


func(a *Auth) Login(ctx context.Context, email string, password string, appId int) (token string, err error){
	const op = "auth.Login"

	log:=a.log.With(slog.String("op", op))

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)

	if err!=nil{
		if errors.Is(err, storage.ErrUserNotFound){
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err:=bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err!=nil{
		a.log.Info("invalid credintials", slog.String("op", op), slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err:=a.appProvider.App(ctx, appId)

	if err!=nil{
		if errors.Is(err, storage.ErrAppNotFound){
			return "", ErrInvalidCredentials
		}
		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	
	token, err= jwt.NewTOken(user, app, a.tokenTTL)
	
	if err!=nil{
		a.log.Error("failed to generate token", slog.String("error", err.Error()))
		
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged succesfully")
	
	return token, nil
}


func(a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error){
	const op = "auth.RegisterNewUser"
	
	log:= a.log.With(slog.String("op", op))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err!=nil{
		log.Error("failed to generate password hash", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err=a.userSaver.SaveUser(ctx, email, passHash)

	if err!=nil{
		if errors.Is(err, storage.ErrUserExists){
			return 0, ErrInvalidCredentials
		}
		log.Error("failed to save user", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return userID, nil
}


func(a *Auth) Role(ctx context.Context, userID int64) (string, error){
	const op = "auth.Role"
	log:=a.log.With(slog.String("op", op))
	
	role, err := a.userProvider.Role(ctx, userID)
	log.Info("get role")

	if err!=nil{
		if errors.Is(err, storage.ErrRoleIsEmpty){
			return "", ErrInvalidCredentials
		}
		log.Error("failed to get role", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("role recieved")

	return role, nil
}

func (a *Auth) RegisterNewApp(ctx context.Context, userID int64, name string, secret string) (int64, error){
	const op = "auth.RegisterNewApp"
	
	log := a.log.With(slog.String("op", op))
	log.Info("register new app")

	role, err := a.Role(ctx, userID)

	if err!=nil{
		if errors.Is(err, storage.ErrRoleIsEmpty){
			return 0, ErrNoRightsForOperation
		}
		log.Error("failed to add app", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if !isAdmin(role){
		return 0, ErrNoRightsForOperation
	}

	appID, err := a.appSaver.SaveApp(ctx, name, secret, userID)

	// TODO
	if err!=nil{
		if errors.Is(err, storage.ErrAppNameExists){
			return 0, ErrInvalidCredentials
		}
		log.Error("failed to add app", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("register has been done")

	return appID, nil
	
}

func isAdmin(role string) bool{
	return role==ADMIN
}