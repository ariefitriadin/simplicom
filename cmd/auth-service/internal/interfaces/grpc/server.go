package grpc

import (
	"context"
	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/access"
	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/persistence/postgres/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	proto "github.com/ariefitriadin/simplicom/cmd/auth-service/proto"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/golang/protobuf/ptypes/empty"
)

type AuthenticationServer struct {
	server                               *server.Server
	queries                              *pgrepo.Queries
	db                                   *pgxpool.Pool
	proto.UnimplementedAuthServiceServer // Embed the unimplemented server
}

func NewServer(server *server.Server, db *pgxpool.Pool) proto.AuthServiceServer {
	return &AuthenticationServer{
		server:  server,
		queries: pgrepo.New(db),
		db:      db,
	}
}

func (s *AuthenticationServer) Register(ctx context.Context, request *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	panic("implement me")
}

func (s *AuthenticationServer) Login(ctx context.Context, request *proto.LoginRequest) (*proto.LoginResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *AuthenticationServer) ValidationBearerToken(ctx context.Context, request *proto.ValidationBearerTokenRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
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
