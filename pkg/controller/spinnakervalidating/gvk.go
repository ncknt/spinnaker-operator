package spinnakervalidating

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func isSpinnakerRequest(req types.Request) bool {
	gv := v1alpha1.SchemeGroupVersion
	return "SpinnakerService" == req.AdmissionRequest.Kind.Kind &&
		gv.Group == req.AdmissionRequest.Kind.Group &&
		gv.Version == req.AdmissionRequest.Kind.Version
}

func isConfigMapRequest(req types.Request) bool {
	gv := corev1.SchemeGroupVersion
	return "ConfigMap" == req.AdmissionRequest.Kind.Kind &&
		gv.Group == req.AdmissionRequest.Kind.Group &&
		gv.Version == req.AdmissionRequest.Kind.Version
}

func (v *spinnakerValidatingController) getSpinnakerService(req types.Request) (*v1alpha1.SpinnakerService, error) {
	if isSpinnakerRequest(req) {
		svc := &v1alpha1.SpinnakerService{}
		if err := v.decoder.Decode(req, svc); err != nil {
			return nil, err
		}
		return svc, nil
	}
	if isConfigMapRequest(req) {
		cm := &corev1.ConfigMap{}
		if err := v.decoder.Decode(req, cm); err != nil {
			return nil, err
		}
		// Check if the configMap is for v spinnaker service
		return v.getMatchedSpinnakerService(cm)
	}
	return nil, nil
}

func (v *spinnakerValidatingController) getSpinnakerServices() ([]v1alpha1.SpinnakerService, error) {
	list := &v1alpha1.SpinnakerServiceList{}
	var opts *client.ListOptions
	ns, _ := k8sutil.GetWatchNamespace()
	if ns == "" {
		opts = &client.ListOptions{}
	} else {
		opts = &client.ListOptions{Namespace: ns}
	}
	err := v.client.List(context.TODO(), opts, list)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (v *spinnakerValidatingController) getMatchedSpinnakerService(cm *corev1.ConfigMap) (*v1alpha1.SpinnakerService, error) {
	ss, err := v.getSpinnakerServices()
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		if s.Spec.SpinnakerConfig.ConfigMap != nil &&
			s.Spec.SpinnakerConfig.ConfigMap.Name == cm.GetName() &&
			s.Spec.SpinnakerConfig.ConfigMap.Namespace == cm.GetNamespace() {
			return &s, nil
		}
	}
	return nil, nil
}
