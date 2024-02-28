package v1beta1

import (
	"context"
	"fmt"

	metav1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRDName int

const (
	Agent CRDName = iota
	Alert
	Single
	Cluster
	Auth
	AlertManager
)

func (c CRDName) String() string {
	return []string{"vmagents.operator.acceldata.io", "vmalerts.operator.acceldata.io", "vmsingles.operator.acceldata.io", "vmclusters.operator.acceldata.io", "vmauths.operator.acceldata.io", "vmalertmanagers.operator.acceldata.io"}[c]
}

type crdInfo struct {
	uuid       types.UID
	kind       string
	apiVersion string
}

var crdCache map[CRDName]*crdInfo

func Init(ctx context.Context, rclient client.Client) error {
	crdCache = make(map[CRDName]*crdInfo)
	var crds metav1.CustomResourceDefinitionList
	if err := rclient.List(ctx, &crds); err != nil {
		return fmt.Errorf("cannot list CRDs during init: %w", err)
	}
	for _, item := range crds.Items {

		var n CRDName
		switch item.Name {
		case "vmagents.operator.acceldata.io":
			n = Agent
		case "vmalerts.operator.acceldata.io":
			n = Alert
		case "vmsingles.operator.acceldata.io":
			n = Single
		case "vmclusters.operator.acceldata.io":
			n = Cluster
		case "vmauths.operator.acceldata.io":
			n = Auth
		case "vmalertmanagers.operator.acceldata.io":
			n = AlertManager
		default:
			continue
		}
		crdCache[n] = &crdInfo{
			uuid:       item.UID,
			apiVersion: metav1.SchemeGroupVersion.String(),
			kind:       "CustomResourceDefinition",
		}
	}
	return nil
}

func GetCRDAsOwner(name CRDName) []v1.OwnerReference {
	crdData := crdCache[name]
	if crdData == nil {
		return nil
	}
	return []v1.OwnerReference{
		{
			Name:       name.String(),
			UID:        crdData.uuid,
			Kind:       "CustomResourceDefinition",
			APIVersion: crdData.apiVersion,
		},
	}
}
