package k8stools

import (
	"testing"

	victoriametricsv1beta1 "github.com/VictoriaMetrics/operator/api/v1beta1"
	"github.com/go-test/deep"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testGetScheme() *runtime.Scheme {
	s := scheme.Scheme
	s.AddKnownTypes(victoriametricsv1beta1.GroupVersion,
		&victoriametricsv1beta1.VMAgent{},
		&victoriametricsv1beta1.VMAgentList{},
		&victoriametricsv1beta1.VMAlert{},
		&victoriametricsv1beta1.VMAlertList{},
		&victoriametricsv1beta1.VMSingle{},
		&victoriametricsv1beta1.VMSingleList{},
		&victoriametricsv1beta1.VMAlertmanager{},
		&victoriametricsv1beta1.VMAlertmanagerList{},
		&victoriametricsv1beta1.VMNodeScrapeList{},
		&victoriametricsv1beta1.VMStaticScrapeList{},
		&victoriametricsv1beta1.VMUserList{},
		&victoriametricsv1beta1.VMAuthList{},
		&victoriametricsv1beta1.VMAlertmanagerConfigList{},
	)
	s.AddKnownTypes(victoriametricsv1beta1.GroupVersion,
		&victoriametricsv1beta1.VMPodScrape{},
		&victoriametricsv1beta1.VMPodScrapeList{},
		&victoriametricsv1beta1.VMServiceScrapeList{},
		&victoriametricsv1beta1.VMServiceScrape{},
		&victoriametricsv1beta1.VMServiceScrapeList{},
		&victoriametricsv1beta1.VMRule{},
		&victoriametricsv1beta1.VMRuleList{},
		&victoriametricsv1beta1.VMProbe{},
		&victoriametricsv1beta1.VMProbeList{},
		&victoriametricsv1beta1.VMNodeScrape{},
		&victoriametricsv1beta1.VMStaticScrape{},
		&victoriametricsv1beta1.VMUser{},
		&victoriametricsv1beta1.VMAuth{},
		&victoriametricsv1beta1.VMAlertmanagerConfig{},
	)
	return s
}

func GetTestClientWithObjects(predefinedObjects []runtime.Object) client.Client {
	obj := []runtime.Object{}
	obj = append(obj, predefinedObjects...)
	fclient := fake.NewFakeClientWithScheme(testGetScheme(), obj...)
	return fclient
}

func CompareObjectMeta(t *testing.T, got, want metav1.ObjectMeta) {
	if diff := deep.Equal(got.Labels, want.Labels); len(diff) > 0 {
		t.Fatalf("objects not match, labels diff: %v", diff)
	}
	if diff := deep.Equal(got.Annotations, want.Annotations); len(diff) > 0 {
		t.Fatalf("objects not match, annotations diff: %v", diff)
	}

	if diff := deep.Equal(got.Name, want.Name); len(diff) > 0 {
		t.Fatalf("objects not match, Name diff: %v", diff)
	}
	if diff := deep.Equal(got.Namespace, want.Namespace); len(diff) > 0 {
		t.Fatalf("objects not match, namespace diff: %v", diff)
	}
}
