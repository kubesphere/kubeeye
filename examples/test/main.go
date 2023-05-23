package main

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
	//klog.Info(get.Spec.Result)
	//var data = make(map[string][]byte)
	//err = json.Unmarshal(get.Spec.Result.Raw, &data)
	//if err != nil {
	//	klog.Error(err)
	//}
	//fmt.Println(rand.Intn(100))
	//data[fmt.Sprintf("node%d", rand.Intn(100))] = s["node1"]
	//marshal, err := json.Marshal(data)
	//get.Spec.Result.Raw = marshal
	//_, err = clients.VersionClientSet.KubeeyeV1alpha2().InspectResults("kubeeye-system").Update(context.TODO(), get, metav1.UpdateOptions{})
	//klog.Error(err)
	//fmt.Println(IsComplete(nil))
}
