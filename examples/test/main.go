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
	//
	//config := []string{"net.ipv4.ip_forward",
	//	"net.bridge.bridge-nf-call-arptables",
	//	"net.bridge.bridge-nf-call-ip6tables",
	//	"net.bridge.bridge-nf-call-iptables",
	//	"net.ipv4.ip_local_reserved_ports",
	//	"vm.max_map_count",
	//	"vm.swappiness",
	//	"vm.overcommit_memory",
	//	"fs.inotify.max_user_instances",
	//	"fs.inotify.max_user_watches",
	//	"kernel.pid_max",
	//	"fs.pipe-max-size",
	//	"net.core.netdev_max_backlog",
	//	"net.core.rmem_max",
	//	"net.core.wmem_max",
	//	"net.ipv4.tcp_max_syn_backlog",
	//	"net.ipv4.neigh.default.gc_thresh1",
	//	"net.ipv4.neigh.default.gc_thresh2",
	//	"net.ipv4.neigh.default.gc_thresh3",
	//	"net.core.somaxconn",
	//	"net.ipv4.conf.all.rp_filter",
	//	"net.ipv4.conf.default.rp_filter",
	//	"net.ipv4.conf.eth0.arp_accept",
	//	"fs.aio-max-nr",
	//	"net.ipv4.tcp_retries2",
	//	"net.ipv4.tcp_max_tw_buckets",
	//	"net.ipv4.tcp_max_orphans",
	//	"net.ipv4.udp_rmem_min",
	//	"net.ipv4.udp_wmem_min"}
	//fmt.Println(config)
	// 获取CPU使用率和空闲率
	//const cpuData = os.cpus();
	//const totalUsage = cpuData.reduce((acc, cur) => acc + cur.times.user + cur.times.sys + cur.times.nice, 0);
	//const totalIdle = cpuData.reduce((acc, cur) => acc + cur.times.idle, 0);
	//const usagePercentage = Math.round(totalUsage / (totalUsage + totalIdle) * 100);
	//const idlePercentage = Math.round(totalIdle / (totalUsage + totalIdle) * 100);
	//fs, err := procfs.NewFS("/proc")
	//if err != nil {
	//
	//}
	//stat, err := fs.Stat()
	//if err != nil {
	//}
	//totalUsage := 0.0
	//totalIdle := 0.0
	//for _, cpuStat := range stat.CPU {
	//	fmt.Println(cpuStat.System)
	//	totalUsage += cpuStat.System + cpuStat.User + cpuStat.Nice
	//	totalIdle += cpuStat.Idle
	//}
	//
	//fmt.Println(totalUsage, totalIdle)
	//fmt.Println(totalUsage / (totalUsage + totalIdle))
	//fmt.Println(totalIdle / (totalUsage + totalIdle))

	//for _, s := range config {
	//	strings, err := fs.SysctlStrings(s)
	//	if err != nil {
	//		klog.Error(err)
	//		continue
	//	}
	//	fmt.Println(strings)
	//
	//}
	//_ = inspect.CSVOutput(clients)

	//if _, err := visitor.CheckRule("etcd = \"active\""); err != nil {
	//	sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
	//	klog.Error(sprintf)
	//} else {
	//	err, res := visitor.EventRuleEvaluate(map[string]interface{}{"etcd": "a"}, "etcd = \"active\"")
	//	if err != nil {
	//		sprintf := fmt.Sprintf("err:%s", err.Error())
	//		klog.Error(sprintf)
	//
	//	} else {
	//		fmt.Println(res)
	//	}
	//
	//}

	//f := excelize.NewFile()
	//defer func() {
	//	if err := f.Close(); err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	//// 创建一个工作表
	//index, err := f.NewSheet("Sheet2")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//// 设置单元格的值
	//f.SetCellValue("Sheet2", "A2", "Hello world.")
	//f.SetSheetRow("Sheet1", "B2", &[]interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9})
	//// 设置工作簿的默认工作表
	//f.SetActiveSheet(index)
	//// 根据指定路径保存文件
	//if err := f.SaveAs("Book1.xlsx"); err != nil {
	//	fmt.Println(err)
	//}
	//file, err := os.Open(path.Join("/proc", "1/mounts"))
	//scanner := bufio.NewScanner(file)
	//defer file.Close()
	//for scanner.Scan() {
	//
	//	fields := strings.Fields(scanner.Text())
	//	fmt.Println(fields)
	//}
	//buf := new(unix.Statfs_t)
	//err = unix.Statfs(path.Join(constant.RootPathPrefix, "/"), buf)
	//if err != nil {
	//
	//}
	//
	//size := float64(buf.Blocks) * float64(buf.Bsize)
	//free := float64(buf.Bfree) * float64(buf.Bsize)
	//avail := float64(buf.Bavail) * float64(buf.Bsize)
	//files := float64(buf.Files)
	//filesFree := float64(buf.Ffree)
	//
	//fmt.Println(size)
	//fmt.Println(free)
	//fmt.Println(avail)
	//fmt.Println(files)
	//fmt.Println(filesFree)

	//client, err := api.NewClient(api.Config{
	//	Address: "http://172.31.73.216:30258",
	//})
	//if err != nil {
	//	klog.Error("create prometheus client failed", err)
	//
	//}
	//queryApi := apiprometheusv1.NewAPI(client)
	//query, _, _ := queryApi.Query(context.TODO(), "node_filesystem_avail_bytes/node_filesystem_size_bytes{mountpoint=~\"/var/lib/docker|/kube|/home|/var|/\"}>0.25", time.Now())
	//marshal, err := json.Marshal(query)
	//
	//var queryResults model.Samples
	//err = json.Unmarshal(marshal, &queryResults)
	//if err != nil {
	//	klog.Error("unmarshal modal Samples failed", err)
	//
	//}
	//var queryResultsMap []map[string]string
	//for i, result := range queryResults {
	//	temp := map[string]string{"value": result.Value.String(), "time": result.Timestamp.String()}
	//	klog.Info(i, result)
	//	for name, value := range result.Metric {
	//		klog.Infof("%s---%s", name, value)
	//		temp[string(name)] = string(value)
	//	}
	//	queryResultsMap = append(queryResultsMap, temp)
	//}

}
