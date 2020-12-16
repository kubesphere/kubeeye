package validator

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conf "kubeye/pkg/config"
	"kubeye/pkg/kube"
	"testing"
)

type Severity string

const (
	// SeverityIgnore ignores validation failures
	SeverityIgnore Severity = "ignore"

	// SeverityWarning warns on validation failures
	SeverityWarning Severity = "warning"
)

type CountSummary struct {
	Successes uint
	Warning   uint
}

func (cs *CountSummary) AddResult(result ResultMessage) {
	if result.Success == false {
		cs.Warning++
	} else {
		cs.Successes++
	}
}

func (rs ResultSet) GetSummary() CountSummary {
	cs := CountSummary{}
	for _, result := range rs {
		cs.AddResult(result)
	}
	return cs
}
func (rs ResultSet) GetWarnings() []ResultMessage {
	warnings := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == conf.SeverityWarning {
			warnings = append(warnings, msg)
		}
	}
	return warnings
}

var resourceConfMinimal = `---
checks:
  cpuLimitsMissing: warning
  livenessProbeMissing: warning
`

func getEmptyWorkload(t *testing.T, name string) kube.GenericWorkload {
	workload, err := kube.NewGenericWorkloadFromPod(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, nil)
	assert.NoError(t, err)
	return workload
}

func testValidate(t *testing.T, container *corev1.Container, resourceConf *string, controllerName string, expectedWarning []ResultMessage, expectedSuccesses []ResultMessage) {
	testValidateWithWorkload(t, container, resourceConf, getEmptyWorkload(t, controllerName), expectedWarning, expectedSuccesses)
}

func testValidateWithWorkload(t *testing.T, container *corev1.Container, resourceConf *string, workload kube.GenericWorkload, expectedWarnings []ResultMessage, expectedSuccesses []ResultMessage) {
	parseConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")
	results, err := applyContainerSchemaChecks(context.Background(), &parseConf, workload, container, false)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(len(expectedWarnings)), results.GetSummary().Warning)
	assert.ElementsMatch(t, expectedWarnings, results.GetWarnings())
}

//Empty config rule
func TestValidateResourceEmptyConfig(t *testing.T) {
	container := &corev1.Container{
		Name: "Empty",
	}
	results, err := applyContainerSchemaChecks(context.Background(), &conf.Configuration{}, getEmptyWorkload(t, ""), container, false)
	if err != nil {
		panic(err)
		assert.Equal(t, 0, results.GetSummary().Successes)
	}
}

func TestValidateResourceEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}
	expectedWarnings := []ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "CPU limits should be set",
			Category: "Resources",
		},
	}

	expectedSuccesses := []ResultMessage{}
	testValidate(t, &container, &resourceConfMinimal, "test", expectedWarnings, expectedSuccesses)
}

func TestValidateHealthChecks(t *testing.T) {

	p3 := map[string]conf.Severity{
		"readinessProbeMissing": conf.SeverityWarning,
		"livenessProbeMissing":  conf.SeverityWarning,
	}

	emptyContainer := &corev1.Container{Name: ""}

	l := ResultMessage{ID: "livenessProbeMissing", Success: false, Severity: "warning", Message: "Liveness probe should be configured", Category: "Health Checks"}
	r := ResultMessage{ID: "readinessProbeMissing", Success: false, Severity: "warning", Message: "Readiness probe should be configured", Category: "Health Checks"}
	f1 := []ResultMessage{r}
	f2 := []ResultMessage{l}

	var testCases = []struct {
		name      string
		probes    map[string]conf.Severity
		container *corev1.Container
		isInit    bool
		dangers   *[]ResultMessage
		warnings  *[]ResultMessage
	}{
		{name: "probes required & not configured", probes: p3, container: emptyContainer, warnings: &f1, dangers: &f2},
	}

	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(context.Background(), &conf.Configuration{Checks: tt.probes}, controller, tt.container, tt.isInit)
			if err != nil {
				panic(err)
			}
			message := fmt.Sprintf("test case %d", idx)

			if tt.warnings != nil && tt.dangers != nil {
				var wdTest = []ResultMessage{}

				warnings := results.GetWarnings()
				assert.Len(t, warnings, 2, message)

				for _, warningTest := range *tt.warnings {
					wdTest = append(wdTest, warningTest)
				}
				for _, dangerTest := range *tt.dangers {
					wdTest = append(wdTest, dangerTest)
				}
				assert.Len(t, warnings, len(wdTest), message)
				assert.ElementsMatch(t, warnings, wdTest, message)
			}
		})
	}
}

func TestValidateImage(t *testing.T) {
	emptyConf := make(map[string]conf.Severity)
	standardConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityWarning,
		"pullPolicyNotAlways": conf.SeverityIgnore,
	}
	strongConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityWarning,
		"pullPolicyNotAlways": conf.SeverityWarning,
	}

	emptyContainer := &corev1.Container{}
	badContainer := &corev1.Container{Image: "test"}
	lessBadContainer := &corev1.Container{Image: "test:latest", ImagePullPolicy: ""}
	goodContainer := &corev1.Container{Image: "test:1.0.0", ImagePullPolicy: "Always"}

	var testCases = []struct {
		name      string
		image     map[string]conf.Severity
		container *corev1.Container
		expected  []ResultMessage
	}{
		{
			name:      "emptyConf + emptyCV",
			image:     emptyConf,
			container: emptyContainer,
			expected:  []ResultMessage{},
		},
		{
			name:      "standardConf + emptyCV",
			image:     standardConf,
			container: emptyContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "warning",
				Category: "Images",
			}},
		},
		{
			name:      "standardConf + badCV",
			image:     standardConf,
			container: badContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "warning",
				Category: "Images",
			}},
		},
		{
			name:      "standardConf + lessBadCV",
			image:     standardConf,
			container: lessBadContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "warning",
				Category: "Images",
			}},
		},
		{
			name:      "strongConf + badCV",
			image:     strongConf,
			container: badContainer,
			expected: []ResultMessage{{
				ID:       "pullPolicyNotAlways",
				Message:  "Image pull policy should be \"Always\"",
				Success:  false,
				Severity: "warning",
				Category: "Images",
			}, {
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "warning",
				Category: "Images",
			}},
		},
		{
			name:      "strongConf + goodCV",
			image:     strongConf,
			container: goodContainer,
			expected:  []ResultMessage{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(context.Background(), &conf.Configuration{Checks: tt.image}, controller, tt.container, false)
			if err != nil {
				panic(err)
			}
			warnings := results.GetWarnings()
			assert.Len(t, warnings, len(tt.expected))
			assert.ElementsMatch(t, warnings, tt.expected)
		})
	}
}

func TestValidateNetworking(t *testing.T) {
	// Test setup.
	emptyConf := make(map[string]conf.Severity)
	standardConf := map[string]conf.Severity{
		"hostPortSet": conf.SeverityWarning,
	}

	emptyContainer := &corev1.Container{Name: ""}
	badContainer := &corev1.Container{
		Ports: []corev1.ContainerPort{{
			ContainerPort: 3000,
			HostPort:      443,
		}},
	}
	goodContainer := &corev1.Container{
		Ports: []corev1.ContainerPort{{
			ContainerPort: 3000,
		}},
	}

	var testCases = []struct {
		name            string
		networkConf     map[string]conf.Severity
		container       *corev1.Container
		expectedResults []ResultMessage
	}{
		{
			name:            "empty ports + empty validation config",
			networkConf:     emptyConf,
			container:       emptyContainer,
			expectedResults: []ResultMessage{},
		},
		{
			name:        "empty ports + standard validation config",
			networkConf: standardConf,
			container:   emptyContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Networking",
			}},
		},
		{
			name:        "empty ports + strong validation config",
			networkConf: standardConf,
			container:   emptyContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Networking",
			}},
		},
		{
			name:            "host ports + empty validation config",
			networkConf:     emptyConf,
			container:       badContainer,
			expectedResults: []ResultMessage{},
		},
		{
			name:        "host ports + standard validation config",
			networkConf: standardConf,
			container:   badContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port should not be configured",
				Success:  false,
				Severity: "warning",
				Category: "Networking",
			}},
		},
		{
			name:        "no host ports + standard validation config",
			networkConf: standardConf,
			container:   goodContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Networking",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(context.Background(), &conf.Configuration{Checks: tt.networkConf}, controller, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, len(tt.expectedResults))
			assert.ElementsMatch(t, messages, tt.expectedResults)
		})
	}
}
