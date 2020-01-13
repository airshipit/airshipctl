package yaml_test

import (
	"bytes"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

// FakeYaml Used to emulate yaml
type FakeYaml struct {
	Key string
}

// FakeErrorDashWriter fake object to simulate errors of writer
type FakeErrorDashWriter struct {
	Err   error
	Match string
}

func (f *FakeErrorDashWriter) Write(b []byte) (int, error) {
	if string(b) == f.Match {
		// arbitrary error from io package
		return 0, f.Err
	}
	return len(b), nil
}

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

func TestWriteOutErrorsWrongYaml(t *testing.T) {
	src := make(chan int)
	var b bytes.Buffer
	assert.Error(t, utilyaml.WriteOut(&b, src))
}

func TestWriteOutErrorsErrorWriter(t *testing.T) {
	// Some easy to match yaml
	fakeYaml := FakeYaml{Key: "value"}
	fakeWriter := &FakeErrorDashWriter{Err: io.ErrUnexpectedEOF}

	strings := []string{
		utilyaml.DotYamlSeparator,
		utilyaml.DashYamlSeparator,
		"Key: value\n",
	}
	for _, str := range strings {
		fakeWriter.Match = str
		assert.Error(t, utilyaml.WriteOut(fakeWriter, fakeYaml))
	}
}
