package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"math"
	"strings"
)

type nodeInfoInspect struct {
}

func init() {
	RuleOperatorMap[constant.NodeInfo] = &nodeInfoInspect{}
}

const excludePath = "/var/lib/docker|/var/lib/kubelet"

func (o *nodeInfoInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var nodeInfos []kubeeyev1alpha2.NodeInfo
	err := json.Unmarshal(jobRule.RunRule, &nodeInfos)
	if err != nil {
		return nil, err
	}
	if nodeInfos == nil {
		return nil, fmt.Errorf("node info rule is empty")
	}

	var jobTemplate *v1.Job
	if nodeInfos[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *nodeInfos[0].NodeName, nil, constant.NodeInfo)
	} else if nodeInfos[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nodeInfos[0].NodeSelector, constant.NodeInfo)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.NodeInfo)
	}

	_, err = clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, err

}

func (o *nodeInfoInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var nodeInfoResult []kubeeyev1alpha2.NodeInfoResultItem

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		fs, err := procfs.NewFS(constant.DefaultProcPath)
		if err != nil {
			return nil, err
		}
		var nodeInfo []kubeeyev1alpha2.NodeInfo
		err = json.Unmarshal(phase.RunRule, &nodeInfo)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}
		for _, info := range nodeInfo {
			ok := false
			resultItem := kubeeyev1alpha2.NodeInfoResultItem{
				Name:  info.Name,
				Level: info.Level,
			}
			switch strings.ToLower(info.Name) {
			case constant.Cpu:
				data := GetCpu(fs)
				resultItem.Value = fmt.Sprintf("%.0f%%", data)
				err, ok = visitor.EventRuleEvaluate(map[string]interface{}{constant.Cpu: data}, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
				}
			case constant.Memory:
				data := GetMemory(fs)
				resultItem.Value = fmt.Sprintf("%.0f%%", data)
				err, ok = visitor.EventRuleEvaluate(map[string]interface{}{constant.Memory: data}, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
				}
			case constant.Filesystem:
				if info.Mount == nil {
					info.Mount = append(info.Mount, "/")
				}
				for _, m := range info.Mount {
					resultItem.FileSystem.Mount = m
					storage, inode := GetFileSystem(m)
					resultItem.Value = fmt.Sprintf("%.0f%%", storage)
					err, ok = visitor.EventRuleEvaluate(map[string]interface{}{constant.Filesystem: storage}, *info.Rule)
					if err != nil {
						resultItem.Value = err.Error()
					}
					resultItem.Assert = ok
					resultItem.FileSystem.Type = constant.Filesystem
					nodeInfoResult = append(nodeInfoResult, resultItem)
					resultItem.Value = fmt.Sprintf("%.0f%%", inode)
					err, ok = visitor.EventRuleEvaluate(map[string]interface{}{constant.Inode: inode}, *info.Rule)
					if err != nil {
						resultItem.Value = err.Error()
					}
					resultItem.Assert = ok
					resultItem.FileSystem.Type = constant.Inode
					nodeInfoResult = append(nodeInfoResult, resultItem)
				}
				continue
			case constant.LoadAvg:
				a, b, c := GetLoadAvg(fs)
				resultItem.Value = fmt.Sprintf("load1:%.0f,load5:%.0f,load15:%.0f", a, b, c)
				var data = make(map[string]interface{})
				data["load1"] = a
				data["load5"] = b
				data["load15"] = c
				err, ok = visitor.EventRuleEvaluate(data, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
				}
			}
			resultItem.Assert = ok
			nodeInfoResult = append(nodeInfoResult, resultItem)
		}
	}

	marshal, err := json.Marshal(nodeInfoResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *nodeInfoInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var nodeInfoResult []kubeeyev1alpha2.NodeInfoResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &nodeInfoResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for _, item := range nodeInfoResult {
		item.NodeName = runNodeName
		resultCr.Spec.NodeInfo = append(resultCr.Spec.NodeInfo, item)
	}

	return resultCr, nil

}

func GetCpu(fs procfs.FS) float64 {
	stat, err := fs.Stat()
	if err != nil {
		klog.Error("failed to get pu info")
		return 0
	}
	totalUsage := 0.0
	totalIdle := 0.0
	for _, cpuStat := range stat.CPU {
		totalUsage += cpuStat.System + cpuStat.User + cpuStat.Nice
		totalIdle += cpuStat.Idle
	}
	usage := totalUsage / (totalUsage + totalIdle)
	return math.Round(usage * 100)

}

func GetMemory(fs procfs.FS) float64 {

	meminfo, err := fs.Meminfo()
	if err != nil {
		klog.Error("failed to get meminfo")
		return 0
	}
	totalMemory := *meminfo.MemTotal
	freeMemory := *meminfo.MemFree + *meminfo.Buffers + *meminfo.Cached
	usedMemory := totalMemory - freeMemory
	memoryUsage := float64(usedMemory) / float64(totalMemory)
	return math.Round(memoryUsage * 100)

}
func GetLoadAvg(fs procfs.FS) (float64, float64, float64) {

	avg, err := fs.LoadAvg()
	if err != nil {
		klog.Error("failed to loadavg")
		return 0, 0, 0
	}

	return avg.Load1, avg.Load5, avg.Load15
}
func GetFileSystem(p string) (float64, float64) {
	u := new(unix.Statfs_t)
	err := unix.Statfs(constant.RootPathPrefix, u)
	if err != nil {
		klog.Error("failed to get filesystem info")
		return 0, 0
	}

	total := float64(u.Blocks) * float64(u.Bsize)
	useBy := float64(u.Bavail) * float64(u.Bsize)
	storageUse := (total - useBy) / total

	inodeUse := u.Files - u.Ffree
	inodeUseRate := float64(inodeUse) / float64(u.Files)
	return math.Round(storageUse * 100), math.Round(inodeUseRate * 100)
}
