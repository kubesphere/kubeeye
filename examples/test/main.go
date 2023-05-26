package main

import "fmt"

func main() {

	//cluster, _ := kube.GetKubeConfigInCluster()
	//var kc kube.KubernetesClient
	//clients, _ := kc.K8SClients(cluster)

	//list, err := clients.ClientSet.CoreV1().ServiceAccounts("kubeeye-system").List(context.Background(), metav1.ListOptions{})
	//klog.Info(list)
	//var resetNum int32 = 5
	//var Parallelism int32 = 3
	//job := v1.Job{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "test",
	//	},
	//	Spec: v1.JobSpec{
	//		BackoffLimit: &resetNum,
	//		Completions:  &resetNum,
	//		Parallelism:  &Parallelism,
	//		Template: corev1.PodTemplateSpec{
	//			ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
	//			Spec: corev1.PodSpec{
	//				Containers: []corev1.Container{{
	//					Name:    "inspect-jobs",
	//					Image:   "jw008/kubeeye:test",
	//					Command: []string{"inspect"},
	//					Args:    []string{"file-change", "--task-name", "inspect"},
	//					VolumeMounts: []corev1.VolumeMount{{
	//						Name:      "root-path",
	//						ReadOnly:  true,
	//						MountPath: "/host",
	//					}},
	//					ImagePullPolicy: "Always",
	//				}},
	//				ServiceAccountName: "",
	//				RestartPolicy:      corev1.RestartPolicyNever,
	//				Volumes: []corev1.Volume{{
	//					Name: "root-path",
	//					VolumeSource: corev1.VolumeSource{
	//						HostPath: &corev1.HostPathVolumeSource{
	//							Path: "/",
	//						},
	//					},
	//				}},
	//			},
	//		},
	//	},
	//}
	//create, err := clients.ClientSet.BatchV1().Jobs("kubeeye-system").Create(context.Background(), &job, metav1.CreateOptions{})
	//
	//if err != nil {
	//	klog.Error(err)
	//}
	//klog.Info(create)

	//result := &kubeeyev1alpha2.InspectResult{
	//	ObjectMeta: metav1.ObjectMeta{Name: "test-result"},
	//	Spec:       kubeeyev1alpha2.InspectResultSpec{Name: "fileChange"},
	//}
	//i, _ := json.Marshal([]string{"master", "node1", "node2", "node3"})
	//s := map[string][]byte{"node1": i}
	//
	//get, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults("kubeeye-system").Get(context.TODO(), result.Name, metav1.GetOptions{})
	//if err != nil {
	//	klog.Error(err)
	//	if kubeErr.IsNotFound(err) {
	//		marshal, err := json.Marshal(s)
	//		klog.Error(err)
	//
	//		ext := runtime.RawExtension{}
	//		ext.Raw = marshal
	//		result.Spec.Result = ext
	//
	//		create, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults("kubeeye-system").Create(context.TODO(), result, metav1.CreateOptions{})
	//		klog.Info(create, err)
	//		os.Exit(1)
	//	}
	//	os.Exit(1)
	//}
	//

	config := []string{"net.ipv4.ip_forward",
		"net.bridge.bridge-nf-call-arptables",
		"net.bridge.bridge-nf-call-ip6tables",
		"net.bridge.bridge-nf-call-iptables",
		"net.ipv4.ip_local_reserved_ports",
		"vm.max_map_count",
		"vm.swappiness",
		"vm.overcommit_memory",
		"fs.inotify.max_user_instances",
		"fs.inotify.max_user_watches",
		"kernel.pid_max",
		"fs.pipe-max-size",
		"net.core.netdev_max_backlog",
		"net.core.rmem_max",
		"net.core.wmem_max",
		"net.ipv4.tcp_max_syn_backlog",
		"net.ipv4.neigh.default.gc_thresh1",
		"net.ipv4.neigh.default.gc_thresh2",
		"net.ipv4.neigh.default.gc_thresh3",
		"net.core.somaxconn",
		"net.ipv4.conf.all.rp_filter",
		"net.ipv4.conf.default.rp_filter",
		"net.ipv4.conf.eth0.arp_accept",
		"fs.aio-max-nr",
		"net.ipv4.tcp_retries2",
		"net.ipv4.tcp_max_tw_buckets",
		"net.ipv4.tcp_max_orphans",
		"net.ipv4.udp_rmem_min",
		"net.ipv4.udp_wmem_min"}
	fmt.Println(config)
}
