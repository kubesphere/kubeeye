//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuditResults) DeepCopyInto(out *AuditResults) {
	*out = *in
	if in.ResultInfos != nil {
		in, out := &in.ResultInfos, &out.ResultInfos
		*out = make([]ResultInfos, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuditResults.
func (in *AuditResults) DeepCopy() *AuditResults {
	if in == nil {
		return nil
	}
	out := new(AuditResults)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInfo) DeepCopyInto(out *ClusterInfo) {
	*out = *in
	if in.NamespacesList != nil {
		in, out := &in.NamespacesList, &out.NamespacesList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInfo.
func (in *ClusterInfo) DeepCopy() *ClusterInfo {
	if in == nil {
		return nil
	}
	out := new(ClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInsight) DeepCopyInto(out *ClusterInsight) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInsight.
func (in *ClusterInsight) DeepCopy() *ClusterInsight {
	if in == nil {
		return nil
	}
	out := new(ClusterInsight)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterInsight) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInsightList) DeepCopyInto(out *ClusterInsightList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterInsight, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInsightList.
func (in *ClusterInsightList) DeepCopy() *ClusterInsightList {
	if in == nil {
		return nil
	}
	out := new(ClusterInsightList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterInsightList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInsightSpec) DeepCopyInto(out *ClusterInsightSpec) {
	*out = *in
	out.Plugins = in.Plugins
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInsightSpec.
func (in *ClusterInsightSpec) DeepCopy() *ClusterInsightSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterInsightSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInsightStatus) DeepCopyInto(out *ClusterInsightStatus) {
	*out = *in
	in.AfterTime.DeepCopyInto(&out.AfterTime)
	in.ClusterInfo.DeepCopyInto(&out.ClusterInfo)
	out.ScoreInfo = in.ScoreInfo
	if in.AuditResults != nil {
		in, out := &in.AuditResults, &out.AuditResults
		*out = make([]AuditResults, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInsightStatus.
func (in *ClusterInsightStatus) DeepCopy() *ClusterInsightStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterInsightStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeBenchState) DeepCopyInto(out *KubeBenchState) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeBenchState.
func (in *KubeBenchState) DeepCopy() *KubeBenchState {
	if in == nil {
		return nil
	}
	out := new(KubeBenchState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NPDState) DeepCopyInto(out *NPDState) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NPDState.
func (in *NPDState) DeepCopy() *NPDState {
	if in == nil {
		return nil
	}
	out := new(NPDState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Plugins) DeepCopyInto(out *Plugins) {
	*out = *in
	out.NPDState = in.NPDState
	out.KubeBenchState = in.KubeBenchState
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Plugins.
func (in *Plugins) DeepCopy() *Plugins {
	if in == nil {
		return nil
	}
	out := new(Plugins)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceInfos) DeepCopyInto(out *ResourceInfos) {
	*out = *in
	if in.ResultItems != nil {
		in, out := &in.ResultItems, &out.ResultItems
		*out = make([]ResultItems, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceInfos.
func (in *ResourceInfos) DeepCopy() *ResourceInfos {
	if in == nil {
		return nil
	}
	out := new(ResourceInfos)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResultInfos) DeepCopyInto(out *ResultInfos) {
	*out = *in
	in.ResourceInfos.DeepCopyInto(&out.ResourceInfos)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResultInfos.
func (in *ResultInfos) DeepCopy() *ResultInfos {
	if in == nil {
		return nil
	}
	out := new(ResultInfos)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResultItems) DeepCopyInto(out *ResultItems) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResultItems.
func (in *ResultItems) DeepCopy() *ResultItems {
	if in == nil {
		return nil
	}
	out := new(ResultItems)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScoreInfo) DeepCopyInto(out *ScoreInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScoreInfo.
func (in *ScoreInfo) DeepCopy() *ScoreInfo {
	if in == nil {
		return nil
	}
	out := new(ScoreInfo)
	in.DeepCopyInto(out)
	return out
}
