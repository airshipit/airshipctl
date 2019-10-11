package yaml_test

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

func TestWriteOut(t *testing.T) {

	// Create some object, that can be marshaled into yaml
	ob := &metav1.ObjectMeta{
		Name:              "RandomName",
		Namespace:         "uniqueNamespace",
		CreationTimestamp: metav1.NewTime(time.Unix(10, 0)),
		Labels: map[string]string{
			"airshiplabel": "airshipit.org",
			"app":          "foo",
		},
	}

	var b bytes.Buffer

	// WriteOut to buffer
	err := utilyaml.WriteOut(&b, ob)
	if err != nil {
		t.Fatalf("Failed to write out yaml: %v", err)
	}

	// Verify result contents
	// TODO (kkalynovskyi) make more reliable tests
	assert.Contains(t, b.String(), ob.Name)
	assert.Contains(t, b.String(), "airshiplabel: airshipit.org")
	assert.Regexp(t, regexp.MustCompile(`^---.*`), b.String())
	assert.Regexp(t, regexp.MustCompile(`.*\.\.\.\n$`), b.String())

	// Create new ObjectMeta for reverse marshaling test.
	var rob metav1.ObjectMeta

	// Check if you can marshal the results of writeout back to the Object
	err = yaml.Unmarshal(b.Bytes(), &rob)
	if err != nil {
		t.Fatalf("Result of write out can not be transformed back into original object: %v", err)
	}

	// Compare original object with reverse marshaled
	assert.Equal(t, ob, &rob)
}
