package transformer

import (
	"encoding/json"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"path/filepath"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct{}

var th = testHelpers{}

func (h *testHelpers) setupTransformer(generator Generator, t *testing.T) (Transformer, *spinnakerv1alpha1.SpinnakerService, *halconfig.SpinnakerConfig) {
	fakeClient := fake.NewFakeClient()
	return h.setupTransformerWithFakeClient(generator, fakeClient, t)
}

func (h *testHelpers) setupTransformerWithFakeClient(generator Generator, fakeClient client.Client, t *testing.T) (Transformer, *spinnakerv1alpha1.SpinnakerService, *halconfig.SpinnakerConfig) {
	spinSvc := h.setupSpinSvc()
	tr, _ := generator.NewTransformer(spinSvc, fakeClient, log.Log.WithName("spinnakerservice"))
	hc := h.setupSpinnakerConfig(t)
	return tr, spinSvc, hc
}

func (h *testHelpers) setupSpinSvc() *spinnakerv1alpha1.SpinnakerService {
	spinSvc := &spinnakerv1alpha1.SpinnakerService{
		Spec: spinnakerv1alpha1.SpinnakerServiceSpec{
			SpinnakerConfig: spinnakerv1alpha1.SpinnakerFileSource{},
			Expose: spinnakerv1alpha1.ExposeConfig{
				Type: "",
				Service: spinnakerv1alpha1.ExposeConfigService{
					Overrides: map[string]spinnakerv1alpha1.ExposeConfigServiceOverrides{},
				},
			},
		},
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{},
	}
	return spinSvc
}

func (h *testHelpers) objectFromJson(fileName string, target interface{}, t *testing.T) {
	fileContents := h.loadJsonFile(fileName, t)
	err := json.Unmarshal([]byte(fileContents), target)
	if err != nil {
		t.Fatal(err)
	}
}

func (h *testHelpers) loadJsonFile(fileName string, t *testing.T) string {
	path := filepath.Join("testdata", fileName) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

func (h *testHelpers) setupSpinnakerConfig(t *testing.T) *halconfig.SpinnakerConfig {
	path := "testdata/halconfig.yml"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hc interface{}
	err = yaml.Unmarshal(bytes, &hc)
	if err != nil {
		t.Fatal(err)
	}

	path = "testdata/profile_gate.yml"
	bytes, err = ioutil.ReadFile(path)
	var profile interface{}
	err = yaml.Unmarshal(bytes, &profile)
	if err != nil {
		t.Fatal(err)
	}
	config := halconfig.SpinnakerConfig{
		HalConfig: hc,
		Profiles:  map[string]interface{}{},
	}
	config.Profiles["gate"] = profile
	return &config
}

func (h *testHelpers) addServiceToGenConfig(gen *generated.SpinnakerGeneratedConfig, svcName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	svc := &corev1.Service{}
	h.objectFromJson(fileName, svc, t)
	gen.Config[svcName] = generated.ServiceConfig{
		Service: svc,
	}
}

func (h *testHelpers) addDeploymentToGenConfig(gen *generated.SpinnakerGeneratedConfig, depName string, fileName string, t *testing.T) {
	if gen.Config == nil {
		gen.Config = make(map[string]generated.ServiceConfig)
	}
	dep := &v1beta2.Deployment{}
	h.objectFromJson(fileName, dep, t)
	gen.Config[depName] = generated.ServiceConfig{
		Deployment: dep,
	}
}

func (h *testHelpers) objToJson(obj interface{}, t *testing.T) string {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

func (h *testHelpers) assertEqualJSON(expected string, actual string, t *testing.T) {
	var expectedObj interface{}
	var actualObj interface{}

	var err error
	err = json.Unmarshal([]byte(expected), &expectedObj)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(actual), &actualObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, reflect.DeepEqual(expectedObj, actualObj))
}
