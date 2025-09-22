package removeannotated

import "k8s.io/apimachinery/pkg/runtime"

func addDefaultingFuncs(schema *runtime.Scheme) error {
	return RegisterDefaults(schema)
}

func SetDefaults_RemoveAnnotated(obj runtime.Object) {
	args := obj.(*RemoveAnnotatedArgs)
	if args.Namespaces == nil {
		args.Namespaces = nil
	}
	if args.LabelSelector == nil {
		args.LabelSelector = nil
	}
	if args.AnnotationKey == "" {
		args.AnnotationKey = "descheduler.alpha.kubernetes.io/evict-me"
	}
}
