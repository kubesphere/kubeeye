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
	"path"

	"strings"
)

type nodeInfoInspect struct {
}

func init() {
	RuleOperatorMap[constant.NodeInfo] = &nodeInfoInspect{}
}

func (o *nodeInfoInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var nodeInfos []kubeeyev1alpha2.NodeInfo
	_ = json.Unmarshal(jobRule.RunRule, &nodeInfos)

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

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
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
				Name: info.Name,
			}
			switch strings.ToLower(info.ResourcesType) {
			case constant.Cpu:
				data := GetCpu(fs)
				resultItem.Value = fmt.Sprintf("%.0f%%", data[constant.Cpu])
				resultItem.ResourcesType.Type = constant.Cpu
				err, ok = visitor.EventRuleEvaluate(data, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
					resultItem.Assert = true
				}
			case constant.Memory:
				data := GetMemory(fs)
				resultItem.Value = fmt.Sprintf("%.0f%%", data[constant.Memory])
				resultItem.ResourcesType.Type = constant.Memory
				err, ok = visitor.EventRuleEvaluate(data, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
					resultItem.Assert = true
				}
			case constant.Filesystem:
				if utils.IsEmptyValue(info.Mount) {
					info.Mount = "/"
				}
				storage := GetFileSystem(info.Mount)
				resultItem.ResourcesType.Type = constant.Filesystem
				resultItem.ResourcesType.Mount = info.Mount
				resultItem.Value = fmt.Sprintf("%.0f%%", storage[constant.Filesystem])
				err, ok = visitor.EventRuleEvaluate(storage, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
					resultItem.Assert = true
				}
			case constant.Inode:
				if utils.IsEmptyValue(info.Mount) {
					info.Mount = "/"
				}
				inodes := GetInodes(info.Mount)
				resultItem.ResourcesType.Type = constant.Inode
				resultItem.ResourcesType.Mount = info.Mount
				resultItem.Value = fmt.Sprintf("%.0f%%", inodes[constant.Inode])
				err, ok = visitor.EventRuleEvaluate(inodes, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
					resultItem.Assert = true
				}
			case constant.LoadAvg:
				loadAvg := GetLoadAvg(fs)
				resultItem.Value = fmt.Sprintf("load1:%.0f,load5:%.0f,load15:%.0f", loadAvg["load1"], loadAvg["load5"], loadAvg["load15"])
				err, ok = visitor.EventRuleEvaluate(loadAvg, *info.Rule)
				if err != nil {
					resultItem.Value = err.Error()
					resultItem.Assert = true
				}
			}
			if ok || resultItem.Assert {
				resultItem.Level = info.Level
				resultItem.Assert = true
			}
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

	for i := range nodeInfoResult {
		nodeInfoResult[i].NodeName = runNodeName
	}
	resultCr.Spec.NodeInfo = append(resultCr.Spec.NodeInfo, nodeInfoResult...)
	return resultCr, nil

}

func GetCpu(fs procfs.FS) map[string]interface{} {
	stat, err := fs.Stat()
	if err != nil {
		klog.Error("failed to get pu info")
		return nil
	}
	totalUsage := 0.0
	totalIdle := 0.0
	for _, cpuStat := range stat.CPU {
		totalUsage += cpuStat.System + cpuStat.User + cpuStat.Nice
		totalIdle += cpuStat.Idle
	}
	usage := totalUsage / (totalUsage + totalIdle)
	return map[string]interface{}{constant.Cpu: usage * 100}

}

func GetMemory(fs procfs.FS) map[string]interface{} {

	meminfo, err := fs.Meminfo()
	if err != nil {
		klog.Error("failed to get meminfo")
		return nil
	}
	totalMemory := *meminfo.MemTotal
	freeMemory := *meminfo.MemFree + *meminfo.Buffers + *meminfo.Cached
	usedMemory := totalMemory - freeMemory
	memoryUsage := float64(usedMemory) / float64(totalMemory)
	return map[string]interface{}{constant.Memory: memoryUsage * 100}

}
func GetLoadAvg(fs procfs.FS) map[string]interface{} {
	avg, err := fs.LoadAvg()
	if err != nil {
		klog.Error("failed to loadavg")
		return nil
	}
	return map[string]interface{}{"load1": avg.Load1, "load5": avg.Load5, "load15": avg.Load15}
}
func GetFileSystem(p string) map[string]interface{} {
	u := new(unix.Statfs_t)
	err := unix.Statfs(path.Join(constant.RootPathPrefix, p), u)
	if err != nil {
		klog.Error("failed to get filesystem info")
		return nil
	}

	total := float64(u.Blocks) * float64(u.Bsize)
	useBy := float64(u.Bavail) * float64(u.Bsize)
	storageUse := (total - useBy) / total

	return map[string]interface{}{constant.Filesystem: storageUse * 100}
}

func GetInodes(p string) map[string]interface{} {
	u := new(unix.Statfs_t)
	err := unix.Statfs(path.Join(constant.RootPathPrefix, p), u)
	if err != nil {
		klog.Error("failed to get filesystem info")
		return nil
	}

	inodeUse := u.Files - u.Ffree
	inodeUseRate := float64(inodeUse) / float64(u.Files)
	return map[string]interface{}{constant.Inode: inodeUseRate * 100}
}
