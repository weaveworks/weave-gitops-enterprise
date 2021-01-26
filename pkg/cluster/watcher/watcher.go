package watcher

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	Added   string = "added"
	Updated string = "updated"
	Deleted string = "deleted"
)

// Function to invoke for every resource object retrieved from the store.
type ProcessFunc = func(eventType string, obj interface{}) error

// Common struct to use when watching a known resource. Every watcher
// created will have its own work queue that decouples delivery of
// an object from its processing. It will also have an Informer that
// is responsible for storing the object for later retrieval and enqueing
// the stored object's key so that it can be processed asychnonously by
// the ProcessFunc function. Based on the workqueue example from client-go
// https://github.com/kubernetes/client-go/blob/master/examples/workqueue/main.go
// but using a SharedIndexInformer instead of an Indexer/Controller
type Watcher struct {
	queue    workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
	process  ProcessFunc
}

// Creates a new Watcher for known resources.
func NewWatcher(informer cache.SharedIndexInformer, process ProcessFunc) *Watcher {

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				event := Event{
					Key:      key,
					Type:     Added,
					Resource: obj,
				}
				log.Debugf("Enqueing %s event for %s", event.Type, event.Key)
				queue.Add(event)
			} else {
				log.Errorf("Failed to enqueue object %T", obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				event := Event{
					Key:      key,
					Type:     Updated,
					Resource: new,
				}
				log.Debugf("Enqueing %s event for %s", event.Type, event.Key)
				queue.Add(event)
			} else {
				log.Errorf("Failed to enqueue object %T", new)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				event := Event{
					Key:      key,
					Type:     Deleted,
					Resource: obj,
				}
				log.Debugf("Enqueing %s event for %s", event.Type, event.Key)
				queue.Add(event)
			} else {
				log.Errorf("Failed to enqueue object %T", obj)
			}
		},
	})

	return &Watcher{
		informer: informer,
		queue:    queue,
		process:  process,
	}
}

func (w *Watcher) runWorker() {
	for w.processNextItem() {
	}
}

func (w *Watcher) processNextItem() bool {
	// Wait until there is a new item in the working queue
	item, quit := w.queue.Get()
	if quit {
		return false
	}

	// Tell the queue that we are done with processing this key. This
	// unblocks the key for other workers. This allows safe parallel
	// processing because two objects with the same key are never processed
	// in parallel.
	defer w.queue.Done(item)

	// Invoke the method containing the business logic
	err := w.processItem(item.(Event))
	// Handle the error if something went wrong during the execution of
	// the business logic.
	w.handleErr(err, item)
	return true
}

// Processes items from the work queue. In case an error happened, it has
// to simply return the error. The retry logic should not be part of this.
func (w *Watcher) processItem(item Event) error {
	// Although we could query the store for the object here
	// we're likely interested in the state of the resource
	// at the time it happened. Querying the store may return
	// more up-to-date information than we would like i.e. an
	// updated resource for an "added" operation.
	return w.process(item.Type, item.Resource)
}

// handleErr checks if an error happened and makes sure we will retry later.
func (w *Watcher) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every
		// successful synchronization.This ensures that future processing
		// of updates for this key is not delayed because of an outdated
		// error history.
		w.queue.Forget(key)
		return
	}

	// Retry 5 times if something goes wrong. After that, stop trying.
	if w.queue.NumRequeues(key) < 5 {
		log.Warnf("Error syncing object %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed again later.
		w.queue.AddRateLimited(key)
		return
	}

	w.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not
	// successfully process this key
	runtime.HandleError(err)
	log.Errorf("Dropping object %q out of the queue: %v", key, err)
}

// Run begins watching and syncing.
func (w *Watcher) Run(name string, threadiness int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done.
	defer w.queue.ShutDown()
	log.Infof("Starting %s watcher.", name)

	go w.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items
	// from the queue is started.
	if !cache.WaitForCacheSync(stopCh, w.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(w.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Infof("Stopping %s watcher.", name)
}
