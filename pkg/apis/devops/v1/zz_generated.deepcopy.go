// +build !ignore_autogenerated

/*
Copyright The Symphony Authors.

*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Migrate) DeepCopyInto(out *Migrate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Migrate.
func (in *Migrate) DeepCopy() *Migrate {
	if in == nil {
		return nil
	}
	out := new(Migrate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Migrate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateCondition) DeepCopyInto(out *MigrateCondition) {
	*out = *in
	in.LastProbeTime.DeepCopyInto(&out.LastProbeTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateCondition.
func (in *MigrateCondition) DeepCopy() *MigrateCondition {
	if in == nil {
		return nil
	}
	out := new(MigrateCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateList) DeepCopyInto(out *MigrateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Migrate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateList.
func (in *MigrateList) DeepCopy() *MigrateList {
	if in == nil {
		return nil
	}
	out := new(MigrateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateSpec) DeepCopyInto(out *MigrateSpec) {
	*out = *in
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.OverrideReplicas != nil {
		in, out := &in.OverrideReplicas, &out.OverrideReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Chart != nil {
		in, out := &in.Chart, &out.Chart
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Releases != nil {
		in, out := &in.Releases, &out.Releases
		*out = make([]*ReleasesConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ReleasesConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateSpec.
func (in *MigrateSpec) DeepCopy() *MigrateSpec {
	if in == nil {
		return nil
	}
	out := new(MigrateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateStatus) DeepCopyInto(out *MigrateStatus) {
	*out = *in
	if in.ReleaseRevision != nil {
		in, out := &in.ReleaseRevision, &out.ReleaseRevision
		*out = make(map[string]int32, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]MigrateCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateStatus.
func (in *MigrateStatus) DeepCopy() *MigrateStatus {
	if in == nil {
		return nil
	}
	out := new(MigrateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleasesConfig) DeepCopyInto(out *ReleasesConfig) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleasesConfig.
func (in *ReleasesConfig) DeepCopy() *ReleasesConfig {
	if in == nil {
		return nil
	}
	out := new(ReleasesConfig)
	in.DeepCopyInto(out)
	return out
}
