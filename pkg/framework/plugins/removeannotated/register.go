package removeannotated

import "k8s.io/apimachinery/pkg/runtime"

var (
	SchemeBuilder      = runtime.NewSchemeBuilder()
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(addDefaultingFuncs)
}
