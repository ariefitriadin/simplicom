package oauth2

import (
	"context"
	"fmt"
	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/config"
	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/persistence/postgres/repositories"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/google/uuid"

	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
	"github.com/ariefitriadin/simplicom/pkg/identity"
	"github.com/ariefitriadin/simplicom/pkg/logger"
	oauth2errors "github.com/go-oauth2/oauth2/v4/errors"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
)

var (
	srv        *oauth2server.Server
	onceServer sync.Once
)

func InitServer(
	cfg *config.Config,
	manager oauth2.Manager,
	timeout time.Duration,
	queries *pgrepo.Queries,
) *oauth2server.Server {
	onceServer.Do(func() {
		srv = oauth2server.NewDefaultServer(manager)

		srv.SetAllowedGrantType(
			oauth2.Refreshing,
			oauth2.ClientCredentials, // Example usage of client credentials

			oauth2.AuthorizationCode,
		)
		srv.SetClientInfoHandler(oauth2server.ClientBasicHandler)

		srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
			i, isAuthorized := identity.FromContext(r.Context())
			if !isAuthorized {
				if r.Form == nil {
					if err := r.ParseForm(); err != nil {
						return "", apperrors.Wrap(err)
					}
				}

				http.Redirect(w, r, fmt.Sprintf("%s?%s", cfg.App.AuthorizeURL, r.Form.Encode()), http.StatusFound)

				return "", nil
			}

			return i.UserID.String(), nil
		})

		srv.SetClientScopeHandler(func(tgr *oauth2.TokenGenerateRequest) (allowed bool, err error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			cid, err := uuid.Parse(tgr.ClientID)
			if err != nil {
				return false, apperrors.Wrap(err)
			}

			client, err := queries.GetClientByID(ctx, cid)
			if err != nil {
				return false, apperrors.Wrap(err) // Return false and the error if there's an issue fetching the client
			}

			tokenScopes := strings.Split(tgr.Scope, ",")
			clientScopes := client.Scope

			if len(tokenScopes) > len(clientScopes) {
				logger.Debug(ctx, fmt.Sprintf("Token not allowed: scopes do not match len(tokenScopes) > len(clientScopes) %v > %v", tokenScopes, clientScopes))
				return false, nil
			}

			if len(tokenScopes) == 0 || len(clientScopes) == 0 {
				logger.Debug(ctx, fmt.Sprintf("Token not allowed: empty scopes len(tokenScopes) == 0 || len(tokenScopes) == 0 %v %v", tokenScopes, clientScopes))
				return false, nil
			}

			clientScopesMap := make(map[string]struct{}, len(clientScopes))
			for _, clientScope := range client.Scope {
				clientScopesMap[clientScope] = struct{}{}
			}

			for _, s := range strings.Split(tgr.Scope, " ") {
				if _, ok := clientScopesMap[s]; !ok {
					return false, nil
				}
			}

			return true, nil
		})

		srv.SetInternalErrorHandler(func(err error) (re *oauth2errors.Response) {
			return &oauth2errors.Response{
				Error: apperrors.Wrap(err),
			}
		})

		srv.SetResponseErrorHandler(func(re *oauth2errors.Response) {
			logger.Error(context.Background(), fmt.Sprintf("[oAuth2|Server] response error: %v", re))
		})

	})

	return srv
}
