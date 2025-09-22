package removeannotated

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

func ValidateRemoveAnnotatedArgs(obj runtime.Object) error {
	args := obj.(*RemoveAnnotatedArgs)
	var allErrs []error

	// At most one of include/exclude can be set
	if args.Namespaces != nil && len(args.Namespaces.Include) > 0 && len(args.Namespaces.Exclude) > 0 {
		allErrs = append(allErrs, fmt.Errorf("only one of Include/Exclude namespaces can be set"))
	}

	if args.LabelSelector != nil {
		if _, err := metav1.LabelSelectorAsSelector(args.LabelSelector); err != nil {
			allErrs = append(allErrs, fmt.Errorf("failed to get label selectors from strategy's params: %+v", err))
		}
	}

	if args.AnnotationKey != "" {
		if err := validateAnnotationKey(args.AnnotationKey); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return utilerrors.NewAggregate(allErrs)
}

func validateAnnotationKey(key string) error {
	if key == "" {
		return fmt.Errorf("annotationKey must be provided")
	}
	if strings.ContainsAny(key, " \t\n\r") {
		return fmt.Errorf("annotationKey must not contain whitespace")
	}
	return nil
}
