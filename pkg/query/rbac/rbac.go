package rbac

// Usually we're checking access for a list of various objects, given
// a principal. To do this efficiently, we generally want to fetch all
// the rules for the principal once, then run each object through the
// rules in turn, to see if the rules allow access.
//
// By "rules", of course, we mean RBAC `PolicyRule` objects. These are
// associated with a principal via `Role` and `RoleBinding`
// objects. We want to defer as much of the logic here to [Kubernetes'
// own RBAC
// code](https://github.com/kubernetes/kubernetes/blob/master/plugin/pkg/auth/authorizer/rbac/rbac.go),
// so we adapt between the internal model here and the methods there.
//
// Specifically:
//
// rbacvalidation.NewDefaultRuleResolver finds the PolicyRule
// objects associated with a particular user, scoped to a cluster.
//
// Then, rbac.RuleAllows (or for a set of rules, rbac.RulesAllow)
// checks whether a rule permits access to an object.

import (
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	k8suser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	rbacv1helpers "k8s.io/kubernetes/pkg/apis/rbac/v1"
	rbacvalidation "k8s.io/kubernetes/pkg/registry/rbac/validation"
	rbacauth "k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

// NewAuthorizer constructs an authorizer with the things that are
// known statically.
func NewAuthorizer(kindToResource map[string]string) *Authorizer {
	return &Authorizer{
		kindToResource: kindToResource,
	}
}

type Authorizer struct {
	kindToResource map[string]string
}

// ObjectAuthorizer constructs an authorization predicate given the
// roles and rolebindings, for the particular cluster and principal.
func (authz *Authorizer) ObjectAuthorizer(roles []models.Role, rolebindings []models.RoleBinding, principal *auth.UserPrincipal, cluster string) func(models.Object) (bool, error) {
	getlist := &clusterRBACGetLister{
		cluster:      cluster,
		roles:        roles,
		rolebindings: rolebindings,
	}
	getlist.init()
	resolver := rbacvalidation.NewDefaultRuleResolver(
		rbacvalidation.RoleGetter(getlist),
		rbacvalidation.RoleBindingLister(getlist),
		rbacvalidation.ClusterRoleGetter(getlist),
		rbacvalidation.ClusterRoleBindingLister(getlist))
	request := &objectAsAttributes{user: principal, kindToResource: authz.kindToResource}
	return func(obj models.Object) (bool, error) {
		request.object = obj
		rules, err := resolver.RulesFor(request.GetUser(), obj.Namespace)
		var ok bool
		for i := range rules {
			if ok = rbacauth.RuleAllows(request, &rules[i]); ok {
				break
			}
		}
		return ok, err
	}
}

// This is a copy of RuleAllows in
// https://github.com/kubernetes/kubernetes/blob/master/plugin/pkg/auth/authorizer/rbac/rbac.go#L178,
// with added printlns. Replace rbacauth.RuleAllows above if you want
// to check how each rule is failing.
func DebugRuleAllows(req authorizer.Attributes, rule *rbacv1.PolicyRule) bool {
	if req.IsResourceRequest() {
		combinedResource := req.GetResource()
		if len(req.GetSubresource()) > 0 {
			combinedResource = req.GetResource() + "/" + req.GetSubresource()
		}

		var ok bool
		if ok = rbacv1helpers.VerbMatches(rule, req.GetVerb()); !ok {
			fmt.Printf("verb does not match: %#v, %s\n", rule.Verbs, req.GetVerb())
			return false
		}

		if ok = rbacv1helpers.APIGroupMatches(rule, req.GetAPIGroup()); !ok {
			fmt.Printf("apiGroup does not match: %#v, %s\n", rule.APIGroups, req.GetAPIGroup())
			return false
		}

		if ok = rbacv1helpers.ResourceMatches(rule, combinedResource, req.GetSubresource()); !ok {
			fmt.Printf("resource does not match: %#v, %s\n", rule.Resources, req.GetResource())
			return false
		}

		if ok = rbacv1helpers.ResourceNameMatches(rule, req.GetName()); !ok {
			fmt.Printf("resourceName does not match: %#v, %s\n", rule.ResourceNames, req.GetName())
			return false
		}

		fmt.Printf("*** rule allows object:\nrule: %#v\nobject: %#v", rule, req)
		return true
	}

	return rbacv1helpers.VerbMatches(rule, req.GetVerb()) &&
		rbacv1helpers.NonResourceURLMatches(rule, req.GetPath())
}

// Adapts between weave-gitops UserPrincipal and Kubernetes
// [user.Info](https://pkg.go.dev/k8s.io/apiserver/pkg/authentication/user#Info)
type principalAsInfo auth.UserPrincipal

func (p *principalAsInfo) GetGroups() []string {
	return p.Groups
}

func (p *principalAsInfo) GetName() string {
	return p.ID
}

func (p *principalAsInfo) GetUID() string {
	return p.ID
}

func (p *principalAsInfo) GetExtra() map[string][]string {
	return nil
}

// Adapts a models.Object to Kubernetes' authorizer.Attributes
// (representing a request)
type objectAsAttributes struct {
	user           *auth.UserPrincipal
	kindToResource map[string]string // map object Kind to resource when asked for the latter

	object models.Object
}

func (o objectAsAttributes) GetUser() k8suser.Info {
	return (*principalAsInfo)(o.user)
}

func (o objectAsAttributes) GetVerb() string {
	return "list" // FIXME unclear! I guess "list" or "get", but what is the model for Weave GitOps?
}

func (o objectAsAttributes) IsReadOnly() bool {
	return true
}

func (o objectAsAttributes) GetNamespace() string {
	return o.object.Namespace
}

func (o objectAsAttributes) GetResource() string {
	resource, ok := o.kindToResource[o.object.Kind]
	if ok {
		return resource
	}
	return ""
}

func (o objectAsAttributes) GetSubresource() string {
	return "" // TODO validate that you can still access the status
}

func (o objectAsAttributes) GetName() string {
	return o.object.Name
}

func (o objectAsAttributes) GetAPIGroup() string {
	return o.object.APIGroup
}

func (o objectAsAttributes) GetAPIVersion() string {
	return o.object.APIVersion
}

func (o objectAsAttributes) IsResourceRequest() bool {
	return true
}

func (o objectAsAttributes) GetPath() string {
	// This is only consulted for non-resource paths; an object always
	// presents as a resource.
	return ""
}

// These are to implement
// https://pkg.go.dev/k8s.io/kubernetes/pkg/registry/rbac/validation#AuthorizationRuleResolver,
// which finds the relevant PolicyRule objects for a user.
//
// TODO these will be called per cluster, per namespace and per name,
// so memoising or even preprocessing would avoid a lot of table
// scans.

type clusterRBACGetLister struct {
	cluster      string
	roles        []models.Role
	rolebindings []models.RoleBinding

	next               int                              // next index to examine
	rolemap            map[string]int                   // name to roles index
	clusterrolebinding []*rbacv1.ClusterRoleBinding     // so we don't go through the whole list again
	rolebindingmap     map[string][]*rbacv1.RoleBinding // namespace to RoleBindings
}

var notfound = errors.New("not found")

func (c *clusterRBACGetLister) init() {
	c.rolebindingmap = map[string][]*rbacv1.RoleBinding{}
	c.rolemap = map[string]int{}
	c.sortRoleBindings()
}

func (c *clusterRBACGetLister) fileRole(i int) {
	key := c.roles[i].Name
	if c.roles[i].Kind == "Role" {
		key = c.roles[i].Namespace + "/" + key
	}
	c.rolemap[key] = i
	c.next = i + 1
}

func (c *clusterRBACGetLister) sortRoleBindings() {
	for i := range c.rolebindings {
		bindingops++
		if c.rolebindings[i].Cluster == c.cluster {
			if c.rolebindings[i].Kind == "RoleBinding" {
				ns := c.rolebindings[i].Namespace
				nsbindings := c.rolebindingmap[ns]
				c.rolebindingmap[ns] = append(nsbindings, makeRoleBinding(&c.rolebindings[i]))
			} else {
				c.clusterrolebinding = append(c.clusterrolebinding, makeClusterRoleBinding(&c.rolebindings[i]))
			}
		}
	}
}

var roleops int
var bindingops int

func resetops() {
	roleops = 0
	bindingops = 0
}

func (c *clusterRBACGetLister) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	if ind, ok := c.rolemap[name]; ok {
		if ind == -1 {
			return nil, notfound
		}
		return makeClusterRole(&c.roles[ind]), nil
	}

	for i := c.next; i < len(c.roles); i++ {
		roleops++
		if c.roles[i].Cluster == c.cluster {
			c.fileRole(i)
			if c.roles[i].Kind == "ClusterRole" && c.roles[i].Name == name {
				return makeClusterRole(&c.roles[i]), nil
			}
		}
	}
	c.next = len(c.roles)
	c.rolemap[name] = -1
	return nil, notfound
}

func (c *clusterRBACGetLister) ListClusterRoleBindings() ([]*rbacv1.ClusterRoleBinding, error) {
	return c.clusterrolebinding, nil
}

func (c *clusterRBACGetLister) GetRole(namespace, name string) (*rbacv1.Role, error) {
	combo := namespace + "/" + name
	if ind, ok := c.rolemap[combo]; ok { // deal with not found!
		if ind == -1 {
			return nil, notfound
		}
		return makeRole(&c.roles[ind]), nil
	}

	for i := c.next; i < len(c.roles); i++ {
		roleops++
		if c.roles[i].Cluster == c.cluster {
			c.fileRole(i)
			if c.roles[i].Kind == "Role" &&
				c.roles[i].Namespace == namespace &&
				c.roles[i].Name == name {
				return makeRole(&c.roles[i]), nil
			}
		}
	}
	c.next = len(c.roles)
	c.rolemap[combo] = -1
	return nil, notfound
}

func (c *clusterRBACGetLister) ListRoleBindings(namespace string) ([]*rbacv1.RoleBinding, error) {
	return c.rolebindingmap[namespace], nil
}

// These are essentially undoing the transformation done in
// internal/models/adapters/.

func makeClusterRole(role *models.Role) *rbacv1.ClusterRole {
	r := &rbacv1.ClusterRole{
		Rules: makePolicyRules(role.PolicyRules),
	}
	r.SetName(role.Name)
	return r
}

func makeRole(role *models.Role) *rbacv1.Role {
	r := &rbacv1.Role{
		Rules: makePolicyRules(role.PolicyRules),
	}
	r.SetNamespace(role.Namespace)
	r.SetName(role.Name)
	return r
}

func makePolicyRules(rules []models.PolicyRule) []rbacv1.PolicyRule {
	rs := make([]rbacv1.PolicyRule, len(rules))
	for i := range rules {
		rs[i] = rbacv1.PolicyRule{
			APIGroups:     models.SplitRuleData(rules[i].APIGroups),
			Resources:     models.SplitRuleData(rules[i].Resources),
			Verbs:         models.SplitRuleData(rules[i].Verbs),
			ResourceNames: models.SplitRuleData(rules[i].ResourceNames),
		}
	}
	return rs
}

func makeClusterRoleBinding(binding *models.RoleBinding) *rbacv1.ClusterRoleBinding {
	b := &rbacv1.ClusterRoleBinding{
		RoleRef: rbacv1.RoleRef{
			Kind: binding.RoleRefKind,
			Name: binding.RoleRefName,
			// FIXME namespace?
		},
		Subjects: makeSubjects(binding.Subjects),
	}
	b.SetName(binding.Name)
	return b
}

func makeRoleBinding(binding *models.RoleBinding) *rbacv1.RoleBinding {
	b := &rbacv1.RoleBinding{
		RoleRef: rbacv1.RoleRef{
			Kind: binding.RoleRefKind,
			Name: binding.RoleRefName,
			// FIXME namespace?
		},
		Subjects: makeSubjects(binding.Subjects),
	}
	b.SetName(binding.Name)
	b.SetNamespace(binding.Namespace)
	return b
}

func makeSubjects(subjects []models.Subject) []rbacv1.Subject {
	ss := make([]rbacv1.Subject, len(subjects))
	for i := range subjects {
		ss[i] = rbacv1.Subject{
			APIGroup:  subjects[i].APIGroup,
			Kind:      subjects[i].Kind,
			Namespace: subjects[i].Namespace,
			Name:      subjects[i].Name,
		}
	}
	return ss
}
