package transformer

import (
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1beta2"
	"testing"
)

func TestTransformManifests_CustomServerPort(t *testing.T) {
	tr, _, hc := th.setupTransformer(&serverPortTransformerGenerator{}, t)
	gen := &generated.SpinnakerGeneratedConfig{}
	th.addDeploymentToGenConfig(gen, "gate", "input_deployment.json", t)

	err := tr.TransformManifests(nil, hc, gen, nil)
	assert.Nil(t, err)

	expected := &v1beta2.Deployment{}
	th.objectFromJson("input_deployment.json", expected, t)
	expected.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = int32(8084)
	expected.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec.Command[4] = "http://localhost:8084/health"
	assert.Equal(t, expected, gen.Config["gate"].Deployment)
}
