package entitlement

import (
	"context"
	_ "embed"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

// LoadEntitlementIntoContextHandler retrieves the entitlement from Kubernetes
// and adds it to the request context.
func EntitlementHandler(ctx context.Context, c client.Client, key client.ObjectKey, next http.Handler) http.Handler {
	var sec v1.Secret
	if err := c.Get(ctx, key, &sec); err != nil {
		log.Warnf("Entitlement cannot be retrieved: %v", err)
		return next
	}

	ent, err := entitlement.VerifyEntitlement(strings.NewReader(public), string(sec.Data["entitlement"]))
	if err != nil {
		log.Warnf("Entitlement was not verified successfully: %v", err)
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyEntitlement, ent)))
	})
}

// CheckEntitlementHandler looks for an entitlement in the request context and
// returns a 500 if the entitlement is not found or appends an HTTP header with
// an expired message.
func CheckEntitlementHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ent, ok := entitlementFromContext(r.Context())
		if ent == nil {
			log.Warnf("Entitlement was not found.")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorMessage))
			return
		}
		if ok {
			if time.Now().After(ent.LicencedUntil) {
				log.Warnf("Entitlement expired on %s.", ent.LicencedUntil.Format("Mon 02 January, 2006"))
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
