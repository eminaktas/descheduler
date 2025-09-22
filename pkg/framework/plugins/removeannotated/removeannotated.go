package removeannotated

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"
	podutil "sigs.k8s.io/descheduler/pkg/descheduler/pod"
	fwtypes "sigs.k8s.io/descheduler/pkg/framework/types"
)

const PluginName = "RemoveAnnotated"

// RemoveAnnotated removes the annotated pods with a specific annotation key on a node.

type RemoveAnnotated struct {
	logger    klog.Logger
	handle    fwtypes.Handle
	args      *RemoveAnnotatedArgs
	podFilter podutil.FilterFunc
}

var _ fwtypes.DeschedulePlugin = &RemoveAnnotated{}

func New(ctx context.Context, args runtime.Object, handle fwtypes.Handle) (fwtypes.Plugin, error) {
	annotationArgs, ok := args.(*RemoveAnnotatedArgs)
	if !ok {
		return nil, fmt.Errorf("args must be of type RemoveAnnotatedArgs, got %T", args)
	}
	logger := klog.FromContext(ctx).WithValues("plugin", PluginName)

	var includedNamespaces, excludedNamespaces sets.Set[string]
	if annotationArgs.Namespaces != nil {
		includedNamespaces = sets.New(annotationArgs.Namespaces.Include...)
		excludedNamespaces = sets.New(annotationArgs.Namespaces.Exclude...)
	}

	// We can combine Filter and PreEvictionFilter since for this strategy it does not matter where we run PreEvictionFilter
	podFilter, err := podutil.NewOptions().
		WithFilter(
			podutil.WrapFilterFuncs(
				handle.Evictor().Filter,
				handle.Evictor().PreEvictionFilter,
			),
		).
		WithNamespaces(includedNamespaces).
		WithoutNamespaces(excludedNamespaces).
		WithLabelSelector(annotationArgs.LabelSelector).
		BuildFilterFunc()
	if err != nil {
		return nil, fmt.Errorf("error initializing pod filter function: %w", err)
	}

	podFilter = podutil.WrapFilterFuncs(podFilter, func(pod *v1.Pod) bool {
		_, exists := pod.Annotations[annotationArgs.AnnotationKey]
		if !exists {
			return false
		}
		return true
	})

	return &RemoveAnnotated{
		logger:    logger,
		handle:    handle,
		args:      annotationArgs,
		podFilter: podFilter,
	}, nil
}

// Name returns the plugin name.
func (d *RemoveAnnotated) Name() string {
	return PluginName
}

// Deschedule extension point implementation for the plugin
func (d *RemoveAnnotated) Deschedule(ctx context.Context, nodes []*v1.Node) *fwtypes.Status {
	logger := klog.FromContext(klog.NewContext(ctx, d.logger)).WithValues("ExtensionPoint", fwtypes.DescheduleExtensionPoint)
	for _, node := range nodes {
		logger.V(2).Info("Processing node", "node", klog.KObj(node))
		pods, err := podutil.ListAllPodsOnANode(node.Name, d.handle.GetPodsAssignedToNodeFunc(), d.podFilter)
		if err != nil {
			// no pods evicted as error encountered retrieving evictable Pods
			return &fwtypes.Status{
				Err: fmt.Errorf("error listing pods on a node: %v", err),
			}
		}
		totalPods := len(pods)
	loop:
		for i := 0; i < totalPods; i++ {
			err := d.handle.Evictor().Evict(ctx, pods[i], evictions.EvictOptions{StrategyName: PluginName})
			if err == nil {
				continue
			}
			switch err.(type) {
			case *evictions.EvictionNodeLimitError:
				continue loop
			case *evictions.EvictionTotalLimitError:
				return nil
			default:
				logger.Error(err, "eviction failed")
			}
		}
	}
	return nil
}
