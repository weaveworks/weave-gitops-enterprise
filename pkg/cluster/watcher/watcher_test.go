package watcher_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/watcher"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestWatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clientset := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events().Informer()

	events := make(chan *v1.Event, 10)

	w := watcher.NewWatcher(informer, func(eventType string, obj interface{}) error {
		event := obj.(*v1.Event)
		events <- event
		return nil
	})

	go w.Run("events", 1, ctx.Done())

	// Don't forget this otherwise the watchers won't fire any event handlers
	cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

	created := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "new"}}
	created, err := clientset.CoreV1().Events("default").Create(context.TODO(), created, metav1.CreateOptions{})
	assert.NoError(t, err)
	created.Labels = map[string]string{"foo": "bar"}
	updated, err := clientset.CoreV1().Events("default").Update(context.TODO(), created, metav1.UpdateOptions{})
	assert.NoError(t, err)
	err = clientset.CoreV1().Events("default").Delete(context.TODO(), updated.Name, metav1.DeleteOptions{})
	assert.NoError(t, err)

	// Stop watcher after 0.5 sec
	time.Sleep(500 * time.Millisecond)
	cancel()

	assert.Equal(t, 3, len(events))
}

func TestWatcherRetry5TimesSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clientset := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events().Informer()

	events := make(chan *v1.Event, 10)

	invoked := &Counter{}
	w := watcher.NewWatcher(informer, func(eventType string, obj interface{}) error {
		invoked.Inc()
		// Return an error the first 5 attempts to process this event
		if invoked.Value() <= 5 {
			return errors.New("oops")
		}

		event := obj.(*v1.Event)
		events <- event
		return nil
	})

	go w.Run("events", 1, ctx.Done())

	// Don't forget this otherwise the watchers won't fire any event handlers
	cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

	created := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "new"}}
	_, err := clientset.CoreV1().Events("default").Create(context.TODO(), created, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Stop watcher after 0.5 sec
	time.Sleep(500 * time.Millisecond)
	cancel()

	assert.Equal(t, 6, invoked.Value())
	assert.Equal(t, 1, len(events))
}

func TestWatcherRetry6TimesFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clientset := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Events().Informer()

	events := make(chan *v1.Event, 10)

	invoked := &Counter{}
	w := watcher.NewWatcher(informer, func(eventType string, obj interface{}) error {
		invoked.Inc()
		// Return an error the first 6 times this event is processed
		if invoked.Value() <= 6 {
			return errors.New("oops")
		}

		event := obj.(*v1.Event)
		events <- event
		return nil
	})

	go w.Run("events", 1, ctx.Done())

	// Don't forget this otherwise the watchers won't fire any event handlers
	cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)

	created := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "new"}}
	_, err := clientset.CoreV1().Events("default").Create(context.TODO(), created, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Stop watcher after 0.5 sec
	time.Sleep(500 * time.Millisecond)
	cancel()

	assert.Equal(t, 6, invoked.Value())
	assert.Equal(t, 0, len(events))
}

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
