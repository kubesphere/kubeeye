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

func TestValidateSecurity(t *testing.T) {
	trueVar := true
	falseVar := false

	// Test setup.
	emptyConf := map[string]conf.Severity{}
	standardConf := map[string]conf.Severity{
		"runAsRootAllowed":           conf.SeverityWarning,
		"runAsPrivileged":            conf.SeverityWarning,
		"notReadOnlyRootFilesystem":  conf.SeverityWarning,
		"privilegeEscalationAllowed": conf.SeverityWarning,
		"dangerousCapabilities":      conf.SeverityWarning,
		"insecureCapabilities":       conf.SeverityWarning,
	}

	emptyContainer := &corev1.Container{Name: ""}
	badContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &falseVar,
			ReadOnlyRootFilesystem:   &falseVar,
			Privileged:               &trueVar,
			AllowPrivilegeEscalation: &trueVar,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"AUDIT_WRITE", "SYS_ADMIN", "NET_ADMIN"},
			},
		},
	}
	emptyPodSpec := &corev1.PodSpec{}
	goodPodSpec := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &trueVar,
		},
	}
	badPodSpec := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}

	goodContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar,
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"NET_BIND_SERVICE", "FOWNER"},
			},
		},
	}

	var testCases = []struct {
		name            string
		securityConf    map[string]conf.Severity
		container       *corev1.Container
		pod             *corev1.PodSpec
		expectedResults []ResultMessage
	}{
		{
			name:            "empty security context + empty validation config",
			securityConf:    emptyConf,
			container:       emptyContainer,
			pod:             emptyPodSpec,
			expectedResults: []ResultMessage{},
		},
		{
			name:         "empty security context + standard validation config",
			securityConf: standardConf,
			container:    emptyContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config",
			securityConf: standardConf,
			container:    badContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config with good settings in podspec",
			securityConf: standardConf,
			container:    badContainer,
			pod:          goodPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config from default set in podspec",
			securityConf: standardConf,
			container:    badContainer,
			pod:          badPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + standard validation config",
			securityConf: standardConf,
			container:    goodContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			workload, err := kube.NewGenericWorkloadFromPod(corev1.Pod{Spec: *tt.pod}, nil)
			assert.NoError(t, err)
			results, err := applyContainerSchemaChecks(context.Background(), &conf.Configuration{Checks: tt.securityConf}, workload, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, len(tt.expectedResults))
			assert.ElementsMatch(t, tt.expectedResults, messages)
		})
	}
}

func TestValidateRunAsRoot(t *testing.T) {
	falseVar := false
	trueVar := true
	nonRootUser := int64(1000)
	rootUser := int64(0)
	config := conf.Configuration{
		Checks: map[string]conf.Severity{
			"runAsRootAllowed": conf.SeverityWarning,
		},
	}

	goodContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: &trueVar,
		},
	}
	badContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}
	inheritContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: nil,
		},
	}
	runAsUserContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &nonRootUser,
		},
	}
	runAsUser0Container := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &rootUser,
		},
	}
	badPod := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}
	runAsUserPod := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &nonRootUser,
		},
	}
	emptyPod := &corev1.PodSpec{}

	testCases := []struct {
		name      string
		container *corev1.Container
		pod       *corev1.PodSpec
		message   ResultMessage
	}{
		{
			name:      "pod=false,container=nil",
			container: inheritContainer,
			pod:       badPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=false,container=true",
			container: goodContainer,
			pod:       badPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=nil,container=runAsUser",
			container: runAsUserContainer,
			pod:       emptyPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=nil",
			container: inheritContainer,
			pod:       runAsUserPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=runAsUser0",
			container: runAsUser0Container,
			pod:       runAsUserPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=false",
			pod:       runAsUserPod,
			container: badContainer,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
	}
	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			workload, err := kube.NewGenericWorkloadFromPod(corev1.Pod{Spec: *tt.pod}, nil)
			assert.NoError(t, err)
			results, err := applyContainerSchemaChecks(context.Background(), &config, workload, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, 1)
			if len(messages) > 0 {
				assert.Equal(t, tt.message, messages[0], fmt.Sprintf("Test case %d failed", idx))
			}
		})
	}
}
