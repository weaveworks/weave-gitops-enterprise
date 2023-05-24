package rbac

import (
	"math/rand"
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// Note on use -- you can run this before and after a change to see
// what effect it has. To run just this:
//
//     go test -run XXX -bench RBAC ./pkg/query/rbac

func Benchmark_RBAC(b *testing.B) {
	// This tries to give an indicative score for how efficient the
	// RBAC authoisation is. The worst case is that no rules match any
	// of the objects, and that could be benchmarked
	// trivially. However, in practice, rules are supposed to _allow_
	// things, so you are usually far from the worst case, and
	// implementations will exploit how rules are written. So we try
	// to make something resembling a real world model, with rules
	// that apply to actual types and objects (though not always the
	// types and objects we're intested in), and using wildcards, and
	// so on.
	//
	// There are some simplifying assumptions, as noted in comments, and:
	// - no cluster role aggregation is used (in practice, I think it's used a lot for baked-in RBAC e..g, cluster-admin)
	// - groups aren't used as subjects, just user names; so long as a realistic number of bindings bind, it doesn't matter how they bind

	// this is used to construct names, and also pick between
	// namespaces, roles and role bindings and so on. This is
	// pseudo-random to mitigate programmer bias, but deterministic so
	// we're testing the same thing each time.
	var localrand = rand.New(rand.NewSource(5))

	randName := func() string {
		chars := make([]byte, 8)
		for i := 0; i < 8; i++ {
			chars[i] = byte(localrand.Intn(26) + 97)
		}
		return string(chars)
	}

	randNameSlice := func(num int) []string {
		names := make([]string, num)
		for i := 0; i < num; i++ {
			names[i] = randName()
		}
		return names
	}

	// these are guesstimates for practical numbers
	const (
		numAPIGroups        = 100 // the total number of API groups (that might have rules explicitly targeting them)
		numKindsPerAPIGroup = 10  // the total number of distinct kinds (that might have rules explicitly targeting them)

		numClusters             = 100 // the number of clusters queried over
		numNamespacesPerCluster = 10  // number of namespaces in each cluster; the objects are assumed to be distributed uniformly amongst the clusters
		numClusterRoles         = 10  // assume these are the same in each cluster
		numClusterRoleBindings  = 100 // number of bindings of cluster roles to subjects (usually not our user)
		numRolesPerCluster      = 100 // make these up and assign randomly to a namespace
		oddsOfBindingUser       = 50  // assume most roles bind to one subject, what is the chance (1/oddsOfBindingUser) it's our user?
		numObjectsPerCluster    = 50  // not the total number of objects in a cluster, just the number with a Kind of interest to us
	)

	// define these ahead of time, because we'll want to make sure
	// some of the rules apply to this user.
	username := randName()
	usergroups := []string{randName(), randName(), randName()}

	type groupKind struct {
		group    string
		kind     string
		resource string
	}
	// these are the ones that will appear in query results
	objectGroupKinds := []groupKind{
		// these only have to be close enough ...
		{"sources.fluxcd.io", "GitRepository", "gitrepositories"},
		{"sources.fluxcd.io", "Bucket", "buckets"},
		{"sources.fluxcd.io", "OCIRepository", "ocirepositories"},
		{"kustomize.fluxcd.io", "Kustomization", "kustomizations"},
		{"helm.fluxcd.io", "HelmRelease", "helmrelease"},
		{"helm.fluxcd.io", "HelmChart", "helmcharts"},
		{"helm.fluxcd.io", "HelmRepository", "helmrepositories"},
	}
	otherGroups := randNameSlice(numAPIGroups)
	groupKinds := []groupKind{}
	for i := range otherGroups {
		kinds := randNameSlice(numKindsPerAPIGroup)
		for j := range kinds {
			groupKinds = append(groupKinds, groupKind{otherGroups[i], kinds[j], randName()})
		}
	}
	groupKinds = append(groupKinds, objectGroupKinds...)

	// this is the model of the clusters and what's in them
	kindToResource := map[string]string{}
	for i := range groupKinds {
		kindToResource[groupKinds[i].kind] = groupKinds[i].resource
	}

	namespaces := randNameSlice(numNamespacesPerCluster)

	// policy rules tend to be one of these shapes:
	// - allow access to everything
	// - allow access to a group of kinds
	// I'm simplifying by assuming almost all rules will include "get"
	// and "list", and by assuming it doesn't matter if rules with
	// ResourceNames never match.
	const oddsOfWildcard = 10      // one out of every ten roles/clusterroles is "allow everything"
	const maxRules = 5             // the most policy rules a role or cluster role can have (when not "allow everything")
	const oddsOfResourceName = 100 // one out of every hundred rules names a particular resource

	makePolicyRules := func() []models.PolicyRule {
		if localrand.Intn(oddsOfWildcard) == 0 {
			var rule models.PolicyRule
			// allow everything
			rule.APIGroups = "*"
			rule.Resources = "*"
			rule.Verbs = "get,list"
			return []models.PolicyRule{rule}
		}

		rules := make([]models.PolicyRule, localrand.Intn(maxRules-1)+1)
		for i := range rules {
			// pick a kind to allow access to
			gk := groupKinds[localrand.Intn(len(groupKinds))]
			rules[i].APIGroups = gk.group // shortcut strings.Join
			rules[i].Resources = gk.resource
			rules[i].Verbs = "list,get"
			if localrand.Intn(oddsOfResourceName) == 0 {
				rules[i].ResourceNames = models.JoinRuleData(randNameSlice(3))
			}
		}
		return rules
	}

	// clusterRoles are the same for every cluster; it's assumed they are paved that way
	clusterRoles := make([]models.Role, numClusterRoles)
	for i := range clusterRoles {
		clusterRoles[i] = models.Role{
			ID:          randName(),
			Kind:        "ClusterRole",
			PolicyRules: makePolicyRules(),
		}
	}
	// cluster role bindings are also the same for each cluster. As a
	// thumb on the scales, we make sure one of them pertains to the
	// user, since these are assigned to every cluster (and it would
	// be boring if none of them matched).
	clusterRoleBindings := make([]models.RoleBinding, numClusterRoles)
	for i := range clusterRoleBindings {
		b := &clusterRoleBindings[i]
		b.Name = randName()
		b.Kind = "ClusterRoleBinding"
		b.RoleRefName = clusterRoles[i].Name
		b.RoleRefKind = "ClusterRole"
		b.Subjects = []models.Subject{
			{
				Kind:     "User",
				Name:     randName(),
				APIGroup: "rbac.authorization.k8s.io",
			},
		}
	}
	clusterRoleBindings[0].Subjects[0].Name = username

	clusters := randNameSlice(numClusters)
	roles := []models.Role{}
	rolebindings := []models.RoleBinding{}

	for _, cluster := range clusters {
		for i := range clusterRoles {
			role := clusterRoles[i]
			role.Cluster = cluster
			roles = append(roles, role)
		}
		for i := range clusterRoleBindings {
			binding := clusterRoleBindings[i]
			binding.Cluster = cluster
			rolebindings = append(rolebindings, binding)
		}

		for i := 0; i < numRolesPerCluster; i++ {
			ns := namespaces[localrand.Intn(numNamespacesPerCluster)]
			role := models.Role{
				Cluster:     cluster,
				Namespace:   ns,
				Kind:        "Role",
				Name:        randName(),
				PolicyRules: makePolicyRules(),
			}
			roles = append(roles, role)
			var user string
			if localrand.Intn(oddsOfBindingUser) == 0 {
				user = username
			} else {
				user = randName()
			}

			binding := models.RoleBinding{
				Cluster:     cluster,
				Namespace:   ns,
				Kind:        "RoleBinding",
				Name:        randName(),
				RoleRefName: role.Name,
				RoleRefKind: "Role",
				Subjects: []models.Subject{
					{
						Kind:     "User",
						Name:     user,
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
			}
			rolebindings = append(rolebindings, binding)
		}
	}

	// assume these arrive in "database order"
	localrand.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })
	localrand.Shuffle(len(rolebindings), func(i, j int) { rolebindings[i], rolebindings[j] = rolebindings[j], rolebindings[i] })

	// fetching the roles and rolebindings is assumed to be a constant
	// cost. A wider benchmark could include that in the loop, since
	// it is certainly something that could improve.

	objects := make([]models.Object, numObjectsPerCluster*numClusters)
	for i := range objects {
		gk := objectGroupKinds[localrand.Intn(len(objectGroupKinds))]
		obj := &objects[i]
		obj.Name = randName()
		obj.Namespace = namespaces[localrand.Intn(numNamespacesPerCluster)]
		obj.Cluster = clusters[localrand.Intn(numClusters)]
		obj.APIGroup = gk.group
		obj.Kind = gk.kind
	}

	user := auth.NewUserPrincipal(auth.ID(username), auth.Groups(usergroups))

	println("Roles", len(roles))
	println("RoleBindings", len(rolebindings))
	println("Objects", len(objects))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// this is constructed once and supplied to the service
		authz := NewAuthorizer(kindToResource)

		count := 0
		// this mimics what the query service does for a request; once
		// a cluster authorizer has been constructed, it's cached,
		// because it won't change. If that mode of use changes, so
		// should this.
		authorizers := map[string]func(models.Object) (bool, error){}

		for j := range objects {
			cluster := objects[j].Cluster
			a, ok := authorizers[cluster]
			if !ok {
				a = authz.ObjectAuthorizer(roles, rolebindings, user, cluster)
				authorizers[cluster] = a
			}
			if ok, _ := a(objects[j]); ok {
				count++
			}
		}
		println("Allowed", count) // if this does not give the same result every run, the benchmark is bogus!
	}
}
