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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ACMECertificateConfig) DeepCopyInto(out *ACMECertificateConfig) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make([]DomainSolverConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ACMECertificateConfig.
func (in *ACMECertificateConfig) DeepCopy() *ACMECertificateConfig {
	if in == nil {
		return nil
	}
	out := new(ACMECertificateConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Certificate) DeepCopyInto(out *Certificate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Certificate.
func (in *Certificate) DeepCopy() *Certificate {
	if in == nil {
		return nil
	}
	out := new(Certificate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Certificate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateCondition) DeepCopyInto(out *CertificateCondition) {
	*out = *in
	if in.LastTransitionTime != nil {
		in, out := &in.LastTransitionTime, &out.LastTransitionTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateCondition.
func (in *CertificateCondition) DeepCopy() *CertificateCondition {
	if in == nil {
		return nil
	}
	out := new(CertificateCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateList) DeepCopyInto(out *CertificateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Certificate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateList.
func (in *CertificateList) DeepCopy() *CertificateList {
	if in == nil {
		return nil
	}
	out := new(CertificateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CertificateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateSpec) DeepCopyInto(out *CertificateSpec) {
	*out = *in
	if in.Organization != nil {
		in, out := &in.Organization, &out.Organization
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(v1.Duration)
		**out = **in
	}
	if in.RenewBefore != nil {
		in, out := &in.RenewBefore, &out.RenewBefore
		*out = new(v1.Duration)
		**out = **in
	}
	if in.DNSNames != nil {
		in, out := &in.DNSNames, &out.DNSNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IPAddresses != nil {
		in, out := &in.IPAddresses, &out.IPAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.IssuerRef = in.IssuerRef
	if in.Usages != nil {
		in, out := &in.Usages, &out.Usages
		*out = make([]KeyUsage, len(*in))
		copy(*out, *in)
	}
	if in.ACME != nil {
		in, out := &in.ACME, &out.ACME
		*out = new(ACMECertificateConfig)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateSpec.
func (in *CertificateSpec) DeepCopy() *CertificateSpec {
	if in == nil {
		return nil
	}
	out := new(CertificateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateStatus) DeepCopyInto(out *CertificateStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]CertificateCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastFailureTime != nil {
		in, out := &in.LastFailureTime, &out.LastFailureTime
		*out = (*in).DeepCopy()
	}
	if in.NotAfter != nil {
		in, out := &in.NotAfter, &out.NotAfter
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateStatus.
func (in *CertificateStatus) DeepCopy() *CertificateStatus {
	if in == nil {
		return nil
	}
	out := new(CertificateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DNS01SolverConfig) DeepCopyInto(out *DNS01SolverConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DNS01SolverConfig.
func (in *DNS01SolverConfig) DeepCopy() *DNS01SolverConfig {
	if in == nil {
		return nil
	}
	out := new(DNS01SolverConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainSolverConfig) DeepCopyInto(out *DomainSolverConfig) {
	*out = *in
	if in.Domains != nil {
		in, out := &in.Domains, &out.Domains
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.SolverConfig.DeepCopyInto(&out.SolverConfig)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainSolverConfig.
func (in *DomainSolverConfig) DeepCopy() *DomainSolverConfig {
	if in == nil {
		return nil
	}
	out := new(DomainSolverConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTP01SolverConfig) DeepCopyInto(out *HTTP01SolverConfig) {
	*out = *in
	if in.IngressClass != nil {
		in, out := &in.IngressClass, &out.IngressClass
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTP01SolverConfig.
func (in *HTTP01SolverConfig) DeepCopy() *HTTP01SolverConfig {
	if in == nil {
		return nil
	}
	out := new(HTTP01SolverConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectReference) DeepCopyInto(out *ObjectReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectReference.
func (in *ObjectReference) DeepCopy() *ObjectReference {
	if in == nil {
		return nil
	}
	out := new(ObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SolverConfig) DeepCopyInto(out *SolverConfig) {
	*out = *in
	if in.HTTP01 != nil {
		in, out := &in.HTTP01, &out.HTTP01
		*out = new(HTTP01SolverConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.DNS01 != nil {
		in, out := &in.DNS01, &out.DNS01
		*out = new(DNS01SolverConfig)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SolverConfig.
func (in *SolverConfig) DeepCopy() *SolverConfig {
	if in == nil {
		return nil
	}
	out := new(SolverConfig)
	in.DeepCopyInto(out)
	return out
}
