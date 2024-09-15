package tests

import (
	"fmt"
	"testing"
	"time"

	ssov1 "github.com/Satori27/grpc-proto/gen/go/sso"
	// grpcauth "github.com/Satori27/sso/internal/grpc"
	// "github.com/Satori27/sso/internal/storage"
	grpcauth "github.com/Satori27/sso/internal/grpc"
	"github.com/Satori27/sso/tests/suite"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID = 1
	appSecret = "secret"
	passDefaultLen = 10

	delta = 1
)


func TestRegister_FailCases(t *testing.T){
	ctx, st :=suite.New(t)

	tests:=[]struct{
		name string
		email string
		password string
		appId int
		expectedErr string
	}{
		{
			name: "Register with Empty Password",
			email: gofakeit.Email(),
			password: "",
			expectedErr: grpcauth.PasswordIsRequiredMsg,
		},
		{
			name: "Register with Empty Email",
			email: "",
			password: generatefakePassword(),
			expectedErr: grpcauth.EmailIsRequiredMsg,
		},
	}

	for _, tt:=range tests{
		t.Run(tt.name, func(t *testing.T){
			_, err:=st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email: tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLogin_FailCases(t *testing.T){
	ctx, st :=suite.New(t)

	tests:=[]struct{
		name string
		email string
		password string
		appId int
		expectedErr string
	}{
		{
			name: "Login with Empty Password",
			email: gofakeit.Email(),
			appId: appID,
			password: "",
			expectedErr: grpcauth.PasswordIsRequiredMsg,
		},
		{
			name: "Login with Empty Email",
			email: "",
			appId: appID,
			password: generatefakePassword(),
			expectedErr: grpcauth.EmailIsRequiredMsg,
		},
		{
			name: "Login with Empty AppID",
			email: gofakeit.Email(),
			password: generatefakePassword(),
			expectedErr: grpcauth.AppIsRequiredMsg,
		},
	}

	for _, tt:=range tests{
		t.Run(tt.name, func(t *testing.T){
			_, err:=st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email: tt.email,
				Password: tt.password,
				AppId: int32(tt.appId),
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestRegisterLogin_Login_InvalidUser(t *testing.T){
	ctx, st := suite.New(t)

	email:=gofakeit.Email()

	password:= generatefakePassword()

	logResp, err:=st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		AppId: 1,
		Email: email,
		Password: password,
	})
	require.Contains(t, err.Error(), grpcauth.InvalidLoginPassword)

	require.Empty(t, logResp.GetToken())
}


func TestRegisterLogin_Login_UserExists(t *testing.T){
	ctx, st := suite.New(t)

	email:=gofakeit.Email()

	password:=generatefakePassword()

	regResp, err:=st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: password,
	})

	require.NoError(t, err)

	require.NotEmpty(t, regResp.GetUserId())

	new_password:=generatefakePassword()

	regRespNew, err:=st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: new_password,
	})

	require.Contains(t, err.Error(), grpcauth.UserExistsMsg)

	require.Empty(t, regRespNew.GetUserId())
}

func TestRegisterLogin_Login_HappyPath(t *testing.T){
	ctx, st:=suite.New(t)

	email:=gofakeit.Email()

	password:=generatefakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email: email,
		Password: password,
		AppId: appID,
	})

	loginTime:=time.Now()

	require.NoError(t, err)

	token:=respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err:=jwt.Parse(token, func(token *jwt.Token)(interface{}, error){
		return []byte(appSecret), nil
	})

	fmt.Println(token)

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), delta)
}

func generatefakePassword()string{
	return gofakeit.Password(true, true, true, true, false,passDefaultLen)
}