package grpc

import (
	"context"
	"time"

	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/access"
	pgrepo "github.com/ariefitriadin/simplicom/cmd/auth-service/internal/persistence/postgres/repositories"
	"github.com/ariefitriadin/simplicom/pkg/auth"
	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
	"github.com/ariefitriadin/simplicom/pkg/identity"
	"github.com/ariefitriadin/simplicom/pkg/logger"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	proto "github.com/ariefitriadin/simplicom/cmd/auth-service/proto"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/golang/protobuf/ptypes/empty"
)

type AuthenticationServer struct {
	server                               *server.Server
	queries                              *pgrepo.Queries
	db                                   *pgxpool.Pool
	authenticator                        auth.Authenticator
	signedMethod                         jwt.SigningMethod
	proto.UnimplementedAuthServiceServer // Embed the unimplemented server
}

func NewServer(server *server.Server, authenticator auth.Authenticator, db *pgxpool.Pool) proto.AuthServiceServer {
	return &AuthenticationServer{
		server:        server,
		queries:       pgrepo.New(db),
		db:            db,
		authenticator: authenticator,
		signedMethod:  jwt.SigningMethodHS512,
	}
}

func (s *AuthenticationServer) RegisterUser(ctx context.Context, request *proto.RegisterUserRequest) (*proto.RegisterUserResponse, error) {
	// check if email already registered
	_, err := s.queries.FindUserByEmail(ctx, request.Email)
	if err != nil && err.Error() != apperrors.ErrNoRows.Error() {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}
	defer tx.Rollback(ctx)

	defaultClientID, err := uuid.Parse("7e0d1a45-54ea-4d74-a55b-1bb6cdf72a50")
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, err
	}

	userId, err := uuid.NewRandom()
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	err = s.queries.WithTx(tx).CreateUser(ctx, pgrepo.CreateUserParams{
		ID:       userId,
		Email:    request.Email,
		Phone:    request.Phone,
		Password: string(hashedPassword),
		Role:     1,
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	var permissions identity.Permission
	permissions = permissions.Add(identity.PermissionTokenRead)
	permissions = permissions.Add(identity.PermissionUserRead)
	permissions = permissions.Add(identity.PermissionUserWrite)

	accessToken, tokenExpiresAt, err := s.generateToken(ctx, userId, defaultClientID, permissions)
	if err != nil {
		return nil, err
	}

	client, err := s.queries.GetClientByID(ctx, defaultClientID)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	err = s.queries.WithTx(tx).CreateToken(ctx, pgrepo.CreateTokenParams{
		AccessToken: accessToken,
		UserID:      userId,
		ClientID:    client.ClientID,
		ExpiresAt:   pgtype.Timestamp{Time: time.Unix(tokenExpiresAt, 0), Valid: true},
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	return &proto.RegisterUserResponse{
		Message:     "User registered successfully",
		AccessToken: accessToken,
	}, nil
}

func (s *AuthenticationServer) UserLogin(ctx context.Context, request *proto.UserLoginRequest) (*proto.UserLoginResponse, error) {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}
	defer tx.Rollback(ctx)

	//check email
	user, err := s.queries.FindUserByEmail(ctx, request.Email)
	if err != nil {
		if err.Error() == apperrors.ErrNoRows.Error() {
			return nil, apperrors.New("user not registered")
		}
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, apperrors.New("invalid password")
	}

	clientID, err := s.queries.GetClientIdByUserId(ctx, user.ID)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	// Generate a new token
	var permissions identity.Permission
	permissions = permissions.Add(identity.PermissionTokenRead)
	permissions = permissions.Add(identity.PermissionUserRead)
	permissions = permissions.Add(identity.PermissionUserWrite)

	accessToken, tokenExpiresAt, err := s.generateToken(ctx, user.ID, clientID, permissions)

	if err != nil {
		return nil, err
	}
	//update token
	err = s.queries.WithTx(tx).UpdateTokenByUserID(ctx, pgrepo.UpdateTokenByUserIDParams{
		AccessToken: accessToken,
		ExpiresAt:   pgtype.Timestamp{Time: time.Unix(tokenExpiresAt, 0), Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	return &proto.UserLoginResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *AuthenticationServer) ValidationBearerToken(ctx context.Context, request *proto.ValidationBearerTokenRequest) (*empty.Empty, error) {
	err := s.authenticator.Verify(request.Token, &auth.Claims{})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}
	return &empty.Empty{}, nil
}

func (s *AuthenticationServer) CreateClient(ctx context.Context, c *proto.CreateClientRequest) (*proto.CreateClientResponse, error) {

	cid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	clientSecret, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	err = s.queries.CreateClient(ctx, pgrepo.CreateClientParams{
		ClientID:     cid,
		ClientSecret: clientSecret,
		Domain:       c.Domain,
		RedirectUrl:  c.RedirectUri,
		Scope:        []string{string(access.ScopeAll)},
	})
	if err != nil {
		return nil, err
	}

	return &proto.CreateClientResponse{
		Message: "Client created successfully",
	}, nil
}

func (s *AuthenticationServer) generateToken(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, permissions identity.Permission) (string, int64, error) {
	i := identity.Identity{
		Permission: permissions,
		UserID:     userID,
		ClientID:   clientID,
	}

	tokenExpiresAt := time.Now().Add(356 * 24 * time.Hour).Unix()
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject:   userID.String(),
			ExpiresAt: tokenExpiresAt,
		},
		Identity: &i,
	}

	token := jwt.NewWithClaims(s.signedMethod, &claims)
	accessToken, err := s.authenticator.Sign(token)
	if err != nil {
		logger.Error(ctx, err.Error())
		return "", 0, apperrors.Wrap(err)
	}

	return accessToken, tokenExpiresAt, nil
}
