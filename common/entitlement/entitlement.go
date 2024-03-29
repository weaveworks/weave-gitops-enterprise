package entitlement

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type contextKey string

func (c contextKey) String() string {
	return "entitlement context key " + string(c)
}

const (
	entitlementExpiredMessageHeader = "Entitlement-Expired-Message"
	expiredMessage                  = "Your entitlement for Weave GitOps Enterprise has expired, please contact sales@weave.works."
	errorMessage                    = "No entitlement was found for Weave GitOps Enterprise. Please contact sales@weave.works."
)

var (
	//go:embed public.pem
	public                string
	contextKeyEntitlement = contextKey("entitlement")
)

// mimics the response of google.golang.org/genproto/googleapis/rpc/status.Status without adding it as dependency
type response struct {
	Message string `json:"message"`
}

// LoadEntitlementIntoContextHandler retrieves the entitlement from Kubernetes
// and adds it to the request context.
func EntitlementHandler(ctx context.Context, log logr.Logger, c client.Client, key client.ObjectKey, next http.Handler) http.Handler {
	var sec v1.Secret
	if err := c.Get(ctx, key, &sec); err != nil {
		log.Error(err, "Entitlement cannot be retrieved")
		return next
	}

	ent, err := entitlement.VerifyEntitlement(strings.NewReader(public), string(sec.Data["entitlement"]))
	if err != nil {
		log.Error(err, "Entitlement was not verified successfully")
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyEntitlement, ent)))
	})
}

// CheckEntitlementHandler looks for an entitlement in the request context and
// returns a 500 if the entitlement is not found or appends an HTTP header with
// an expired message.
func CheckEntitlementHandler(log logr.Logger, next http.Handler, publicRoutes []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsPublicRoute(r.URL, publicRoutes) {
			next.ServeHTTP(w, r)
			return
		}
		ent, ok := entitlementFromContext(r.Context())
		if ent == nil {
			log.Info("Entitlement was not found.")
			w.WriteHeader(http.StatusInternalServerError)
			response, err := json.Marshal(
				response{
					Message: errorMessage,
				},
			)
			if err != nil {
				log.Error(err, "unexpected error while handling entitlement not found response")
			}
			w.Write(response)
			return
		}
		if ok {
			if time.Now().After(ent.LicencedUntil) {
				log.Info("Entitlement expired.", "licencedUntil", ent.LicencedUntil.Format("Mon 02 January, 2006"))
				w.Header().Add(entitlementExpiredMessageHeader, expiredMessage)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func entitlementFromContext(ctx context.Context) (*entitlement.Entitlement, bool) {
	ent, ok := ctx.Value(contextKeyEntitlement).(*entitlement.Entitlement)
	return ent, ok
}

func IsPublicRoute(u *url.URL, publicRoutes []string) bool {
	for _, pr := range publicRoutes {
		if u.Path == pr {
			return true
		}
	}

	return false
}
