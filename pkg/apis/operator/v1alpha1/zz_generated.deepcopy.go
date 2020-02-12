// +build !ignore_autogenerated

//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metering) DeepCopyInto(out *Metering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metering.
func (in *Metering) DeepCopy() *Metering {
	if in == nil {
		return nil
	}
	out := new(Metering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Metering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringList) DeepCopyInto(out *MeteringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Metering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringList.
func (in *MeteringList) DeepCopy() *MeteringList {
	if in == nil {
		return nil
	}
	out := new(MeteringList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUI) DeepCopyInto(out *MeteringMultiCloudUI) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUI.
func (in *MeteringMultiCloudUI) DeepCopy() *MeteringMultiCloudUI {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringMultiCloudUI) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUIList) DeepCopyInto(out *MeteringMultiCloudUIList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MeteringMultiCloudUI, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUIList.
func (in *MeteringMultiCloudUIList) DeepCopy() *MeteringMultiCloudUIList {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUIList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringMultiCloudUIList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUISpec) DeepCopyInto(out *MeteringMultiCloudUISpec) {
	*out = *in
	out.MongoDB = in.MongoDB
	out.External = in.External
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUISpec.
func (in *MeteringMultiCloudUISpec) DeepCopy() *MeteringMultiCloudUISpec {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUISpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUISpecExternal) DeepCopyInto(out *MeteringMultiCloudUISpecExternal) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUISpecExternal.
func (in *MeteringMultiCloudUISpecExternal) DeepCopy() *MeteringMultiCloudUISpecExternal {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUISpecExternal)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUISpecMongoDB) DeepCopyInto(out *MeteringMultiCloudUISpecMongoDB) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUISpecMongoDB.
func (in *MeteringMultiCloudUISpecMongoDB) DeepCopy() *MeteringMultiCloudUISpecMongoDB {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUISpecMongoDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringMultiCloudUIStatus) DeepCopyInto(out *MeteringMultiCloudUIStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringMultiCloudUIStatus.
func (in *MeteringMultiCloudUIStatus) DeepCopy() *MeteringMultiCloudUIStatus {
	if in == nil {
		return nil
	}
	out := new(MeteringMultiCloudUIStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSender) DeepCopyInto(out *MeteringSender) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSender.
func (in *MeteringSender) DeepCopy() *MeteringSender {
	if in == nil {
		return nil
	}
	out := new(MeteringSender)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringSender) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSenderList) DeepCopyInto(out *MeteringSenderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MeteringSender, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSenderList.
func (in *MeteringSenderList) DeepCopy() *MeteringSenderList {
	if in == nil {
		return nil
	}
	out := new(MeteringSenderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringSenderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSenderSpec) DeepCopyInto(out *MeteringSenderSpec) {
	*out = *in
	out.Sender = in.Sender
	out.MongoDB = in.MongoDB
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSenderSpec.
func (in *MeteringSenderSpec) DeepCopy() *MeteringSenderSpec {
	if in == nil {
		return nil
	}
	out := new(MeteringSenderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSenderSpecMongoDB) DeepCopyInto(out *MeteringSenderSpecMongoDB) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSenderSpecMongoDB.
func (in *MeteringSenderSpecMongoDB) DeepCopy() *MeteringSenderSpecMongoDB {
	if in == nil {
		return nil
	}
	out := new(MeteringSenderSpecMongoDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSenderSpecSender) DeepCopyInto(out *MeteringSenderSpecSender) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSenderSpecSender.
func (in *MeteringSenderSpecSender) DeepCopy() *MeteringSenderSpecSender {
	if in == nil {
		return nil
	}
	out := new(MeteringSenderSpecSender)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSenderStatus) DeepCopyInto(out *MeteringSenderStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSenderStatus.
func (in *MeteringSenderStatus) DeepCopy() *MeteringSenderStatus {
	if in == nil {
		return nil
	}
	out := new(MeteringSenderStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSpec) DeepCopyInto(out *MeteringSpec) {
	*out = *in
	out.MongoDB = in.MongoDB
	out.External = in.External
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSpec.
func (in *MeteringSpec) DeepCopy() *MeteringSpec {
	if in == nil {
		return nil
	}
	out := new(MeteringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSpecExternal) DeepCopyInto(out *MeteringSpecExternal) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSpecExternal.
func (in *MeteringSpecExternal) DeepCopy() *MeteringSpecExternal {
	if in == nil {
		return nil
	}
	out := new(MeteringSpecExternal)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringSpecMongoDB) DeepCopyInto(out *MeteringSpecMongoDB) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringSpecMongoDB.
func (in *MeteringSpecMongoDB) DeepCopy() *MeteringSpecMongoDB {
	if in == nil {
		return nil
	}
	out := new(MeteringSpecMongoDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringStatus) DeepCopyInto(out *MeteringStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringStatus.
func (in *MeteringStatus) DeepCopy() *MeteringStatus {
	if in == nil {
		return nil
	}
	out := new(MeteringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUI) DeepCopyInto(out *MeteringUI) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUI.
func (in *MeteringUI) DeepCopy() *MeteringUI {
	if in == nil {
		return nil
	}
	out := new(MeteringUI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringUI) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUIList) DeepCopyInto(out *MeteringUIList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MeteringUI, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUIList.
func (in *MeteringUIList) DeepCopy() *MeteringUIList {
	if in == nil {
		return nil
	}
	out := new(MeteringUIList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeteringUIList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUISpec) DeepCopyInto(out *MeteringUISpec) {
	*out = *in
	out.MongoDB = in.MongoDB
	out.External = in.External
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUISpec.
func (in *MeteringUISpec) DeepCopy() *MeteringUISpec {
	if in == nil {
		return nil
	}
	out := new(MeteringUISpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUISpecExternal) DeepCopyInto(out *MeteringUISpecExternal) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUISpecExternal.
func (in *MeteringUISpecExternal) DeepCopy() *MeteringUISpecExternal {
	if in == nil {
		return nil
	}
	out := new(MeteringUISpecExternal)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUISpecMongoDB) DeepCopyInto(out *MeteringUISpecMongoDB) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUISpecMongoDB.
func (in *MeteringUISpecMongoDB) DeepCopy() *MeteringUISpecMongoDB {
	if in == nil {
		return nil
	}
	out := new(MeteringUISpecMongoDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeteringUIStatus) DeepCopyInto(out *MeteringUIStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeteringUIStatus.
func (in *MeteringUIStatus) DeepCopy() *MeteringUIStatus {
	if in == nil {
		return nil
	}
	out := new(MeteringUIStatus)
	in.DeepCopyInto(out)
	return out
}
