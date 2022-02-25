package dashboard

/*func GetInfo(ctx context.Context) *OverviewResponse {
	// get kubernetes resources and put into the channel.
	go func(ctx context.Context, kubeconfig string) {
		err := kube.GetK8SResourcesProviderForOverview(ctx, "")
		if err != nil {
			panic(err)
		}
	}(ctx ,"")
	overview := &OverviewResponse{
		ApiVersion: overviewApi,
		Kind: overviewKind,
		Metadata: Metadata{
			name,
		},
	}
	k8sResources := <-kube.K8sOverviewResourcesChan
	nsList := []string{}
	for _, ns := range k8sResources.Namespaces.Items {
		nsList = append(nsList,ns.GetName())
	}
	workloadsCount := len(k8sResources.Deployments.Items) + len(k8sResources.DaemonSets.Items) + len(k8sResources.CronJobs.Items) + len(k8sResources.Jobs.Items) + len(k8sResources.StatefulSets.Items)
	overview.Spec = OverviewSpec{
		ClusterVersion:  k8sResources.ServerVersion.GitVersion,
		NodesCount:      len(k8sResources.Nodes.Items),
		NamespacesCount: len(k8sResources.Namespaces.Items),
		WorkloadsCount:  workloadsCount,
		NamespacesList:  nsList,
	}
	fmt.Println(overview)
	return overview
}*/