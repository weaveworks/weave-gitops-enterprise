package utils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/aggregator"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/collector"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"
)

// StatusChecker is a wrapper around the StatusPoller
// that provides a way to poll the status of a set of resources
type StatusChecker struct {
	pollInterval time.Duration
	timeout      time.Duration
	client       client.Client
	statusPoller *polling.StatusPoller
	logger       logger.Logger
}

// NewStatusChecker returns a new StatusChecker that will use the provided client to poll the status of the resources
func NewStatusChecker(client k8s_client.Client, pollInterval time.Duration, timeout time.Duration, log logger.Logger) (*StatusChecker, error) {

	return &StatusChecker{
		pollInterval: pollInterval,
		timeout:      timeout,
		client:       client,
		statusPoller: polling.NewStatusPoller(client, client.RESTMapper(), polling.Options{}),
		logger:       log,
	}, nil
}

// Assess will poll the status of the provided resources until all resources have reached the desired status
func (sc *StatusChecker) Assess(identifiers ...object.ObjMetadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), sc.timeout)
	defer cancel()

	opts := polling.PollOptions{PollInterval: sc.pollInterval}
	eventsChan := sc.statusPoller.Poll(ctx, identifiers, opts)

	coll := collector.NewResourceStatusCollector(identifiers)
	done := coll.ListenWithObserver(eventsChan, desiredStatusNotifierFunc(cancel, status.CurrentStatus))

	<-done

	// we use sorted identifiers to loop over the resource statuses because a Go's map is unordered.
	// sorting identifiers by object's name makes sure that the logs look stable for every run
	sort.SliceStable(identifiers, func(i, j int) bool {
		return strings.Compare(identifiers[i].Name, identifiers[j].Name) < 0
	})
	for _, id := range identifiers {
		rs := coll.ResourceStatuses[id]
		switch rs.Status {
		case status.CurrentStatus:
			sc.logger.Successf("%s: %s ready", rs.Identifier.Name, strings.ToLower(rs.Identifier.GroupKind.Kind))
		case status.NotFoundStatus:
			sc.logger.Failuref("%s: %s not found", rs.Identifier.Name, strings.ToLower(rs.Identifier.GroupKind.Kind))
		default:
			sc.logger.Failuref("%s: %s not ready", rs.Identifier.Name, strings.ToLower(rs.Identifier.GroupKind.Kind))
		}
	}

	if coll.Error != nil || ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timed out waiting for condition")
	}
	return nil
}

// desiredStatusNotifierFunc returns an Observer function for the
// ResourceStatusCollector that will cancel the context (using the cancelFunc)
// when all resources have reached the desired status.
func desiredStatusNotifierFunc(cancelFunc context.CancelFunc,
	desired status.Status) collector.ObserverFunc {
	return func(rsc *collector.ResourceStatusCollector, _ event.Event) {
		var rss []*event.ResourceStatus
		for _, rs := range rsc.ResourceStatuses {
			rss = append(rss, rs)
		}
		aggStatus := aggregator.AggregateStatus(rss, desired)
		if aggStatus == desired {
			cancelFunc()
		}
	}
}
