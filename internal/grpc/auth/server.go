package auth

import (
	"context"
	"errors"

	ssov1 "github.com/Satori27/grpc-proto/gen/go/sso"
	grpcauth "github.com/Satori27/sso/internal/grpc"
	"github.com/Satori27/sso/internal/services/auth"

	// validator "github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type Auth interface{
	Login(ctx context.Context, email string, password string, appId int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	Role(ctx context.Context, userID int64) (string, error)
	RegisterNewApp(ctx context.Context, userID int64, name string, secret string) (int64, error)
}


type serverAPI struct{
	ssov1.UnimplementedAuthServer
	auth Auth
}


func Register(gRPC *grpc.Server, auth Auth){
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}


func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error){
	if req.GetEmail()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.EmailIsRequiredMsg)
	}
	
	if req.GetPassword()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.PasswordIsRequiredMsg)
	}

	if req.GetAppId()==0{
		return nil, status.Error(codes.InvalidArgument, grpcauth.AppIsRequiredMsg)
	}

	token, err:=s.auth.Login(ctx, req.Email, req.Password, int(req.GetAppId()))
	if err!=nil{
		if errors.Is(err, auth.ErrInvalidCredentials){
			return nil, status.Error(codes.InvalidArgument, grpcauth.InvalidLoginPassword)
		}
		
		return nil, status.Error(codes.Internal, grpcauth.InternalError)
	}

	return &ssov1.LoginResponse{Token: token}, nil
}


func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error){
	if req.GetEmail()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.EmailIsRequiredMsg)
	}

	if req.GetPassword()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.PasswordIsRequiredMsg)
	}

	userID, err:=s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())

	if err!=nil{
		if errors.Is(err, auth.ErrInvalidCredentials){
			return nil, status.Error(codes.AlreadyExists, grpcauth.UserExistsMsg)
		}
		return nil, status.Error(codes.Internal, grpcauth.InternalError)
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}


func (s *serverAPI) Role(ctx context.Context, req *ssov1.RoleRequest) (*ssov1.RoleResponse, error){
	if req.GetUserId()==0{
		return nil, status.Error(codes.InvalidArgument, grpcauth.UserIsReuired)
	}

	role, err:=s.auth.Role(ctx, req.UserId)

	if err!=nil{
		if errors.Is(err, auth.ErrInvalidCredentials){
			return nil, status.Error(codes.InvalidArgument, grpcauth.EmptyRole)
		}
		return nil, status.Error(codes.Internal, grpcauth.InternalError)
	}
	return &ssov1.RoleResponse{Role: role}, nil
}

func (s *serverAPI) CreateApp(ctx context.Context, req *ssov1.CreateAppRequest) (*ssov1.CreateAppResponse, error){
	if req.GetUserId()==0{
		return nil, status.Error(codes.InvalidArgument, grpcauth.UserIsReuired)
	}
	if req.GetName()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.AppNameIsRequired)
	}
	if req.GetSecret()==""{
		return nil, status.Error(codes.InvalidArgument, grpcauth.AppSecretIsRequired)
	}

	resp, err := s.auth.RegisterNewApp(ctx, req.GetUserId(), req.Name, req.Secret)
	// TODO:
	if err!=nil{
		if errors.Is(err, auth.ErrInvalidCredentials){
			return &ssov1.CreateAppResponse{}, status.Error(codes.AlreadyExists, grpcauth.AppExistsMsg)
		}
		if errors.Is(err, auth.ErrNoRightsForOperation){
			return &ssov1.CreateAppResponse{}, status.Error(codes.PermissionDenied, "you have no permission for this operation")
		}

		return &ssov1.CreateAppResponse{}, status.Error(codes.Internal, grpcauth.InternalError)
	}

	return &ssov1.CreateAppResponse{AppId: resp}, nil
}