package removeannotated

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/descheduler/pkg/api"
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RemoveAnnotatedArgs struct {
	metav1.TypeMeta `json:",inline"`

	Namespaces    *api.Namespaces       `json:"namespaces,omitempty"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	// AnnotationKey is the pod annotation key that marks a workload for eviction.
	AnnotationKey string `json:"annotationKey"`
}
