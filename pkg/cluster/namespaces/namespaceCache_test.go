package namespaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewNamespaceCache(t *testing.T) {
	cli := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(cli, 0)
	want := NamespacesInformerCache{namespacesInformer: factory.Core().V1().Namespaces()}
	got, err := NewNamespacesInformerCache(factory)
	assert.NoError(t, err)
	assert.Equal(t, want.namespacesInformer.Informer(), got.namespacesInformer.Informer())
}

// func TestNamespacesCacheList(t *testing.T) {
// 	cli := fake.NewSimpleClientset()
// 	watcherStarted := make(chan struct{})
// 	cli.PrependWatchReactor("namespaces", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
// 		gvr := action.GetResource()
// 		ns := action.GetNamespace()
// 		watch, err := cli.Tracker().Watch(gvr, ns)
// 		if err != nil {
// 			return false, nil, err
// 		}
// 		close(watcherStarted)
// 		return true, watch, nil
// 	})
// 	factory := informers.NewSharedInformerFactory(cli, 0)
// 	n := NewNamespaceCache(factory)
// 	stop := make(chan struct{})
// 	factory.Start(stop)

// 	cache.WaitForCacheSync(stop, n.CacheSync())

// 	<-watcherStarted
// 	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}}
// 	_, err := cli.CoreV1().Namespaces().Create(context.Background(), &ns, metav1.CreateOptions{})
// 	assert.NoError(t, err)

// 	nsList, err := n.List()
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, len(nsList))
// 	assert.Equal(t, nsList[0].Name, ns.Name)

// }
