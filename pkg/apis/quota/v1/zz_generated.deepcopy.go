//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2024 Kubeworkz Authors

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeResourceQuota) DeepCopyInto(out *KubeResourceQuota) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeResourceQuota.
func (in *KubeResourceQuota) DeepCopy() *KubeResourceQuota {
	if in == nil {
		return nil
	}
	out := new(KubeResourceQuota)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeResourceQuota) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeResourceQuotaList) DeepCopyInto(out *KubeResourceQuotaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KubeResourceQuota, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeResourceQuotaList.
func (in *KubeResourceQuotaList) DeepCopy() *KubeResourceQuotaList {
	if in == nil {
		return nil
	}
	out := new(KubeResourceQuotaList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeResourceQuotaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeResourceQuotaSpec) DeepCopyInto(out *KubeResourceQuotaSpec) {
	*out = *in
	if in.Hard != nil {
		in, out := &in.Hard, &out.Hard
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	out.Target = in.Target
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeResourceQuotaSpec.
func (in *KubeResourceQuotaSpec) DeepCopy() *KubeResourceQuotaSpec {
	if in == nil {
		return nil
	}
	out := new(KubeResourceQuotaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeResourceQuotaStatus) DeepCopyInto(out *KubeResourceQuotaStatus) {
	*out = *in
	if in.Hard != nil {
		in, out := &in.Hard, &out.Hard
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Used != nil {
		in, out := &in.Used, &out.Used
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.SubResourceQuotas != nil {
		in, out := &in.SubResourceQuotas, &out.SubResourceQuotas
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeResourceQuotaStatus.
func (in *KubeResourceQuotaStatus) DeepCopy() *KubeResourceQuotaStatus {
	if in == nil {
		return nil
	}
	out := new(KubeResourceQuotaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetObj) DeepCopyInto(out *TargetObj) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetObj.
func (in *TargetObj) DeepCopy() *TargetObj {
	if in == nil {
		return nil
	}
	out := new(TargetObj)
	in.DeepCopyInto(out)
	return out
}
