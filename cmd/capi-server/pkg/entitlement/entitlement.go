package entitlement

import (
	"context"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type key int

const (
	entitlementKey                  key = 0
	entitlementExpiredMessageHeader     = "Entitlement-Expired-Message"
)

var public = `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEA140z8yf4+R9MQwwS6yTrWIl/1IBOjLVvh9x87Wd84TU=
-----END PUBLIC KEY-----`

// LoadEntitlementIntoContextHandler retrieves the entitlement from Kubernetes
// and adds it to the request context.
func LoadEntitlementIntoContextHandler(ctx context.Context, c client.Client, key client.ObjectKey, next http.Handler) http.Handler {
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
		ctx := r.Context()
		ctx = context.WithValue(ctx, entitlementKey, ent)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// CheckEntitlementHandler looks for an entitlement in the request context and
// returns a 500 if the entitlement is not found or appends an HTTP header with
// an expired message.
func CheckEntitlementHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ent := ctx.Value(entitlementKey)
		if ent == nil {
			log.Warnf("Entitlement was not found.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
		if e, ok := ent.(*entitlement.Entitlement); ok {
			if time.Now().After(e.LicencedUntil) {
				log.Warnf("Entitlement expired on %s.", e.LicencedUntil.Format("Mon 02 January, 2006"))
				w.Header().Add(entitlementExpiredMessageHeader, "Your entitlement has expired, please contact WeaveWorks")
			}
		}
	})
}
