//
// This file holds the protobuf definitions for messages and enums
// used in the Weave GitOps Terraform API.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: api/terraform/types.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SourceRef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApiVersion string `protobuf:"bytes,1,opt,name=api_version,json=apiVersion,proto3" json:"api_version,omitempty"`
	Kind       string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Name       string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Namespace  string `protobuf:"bytes,4,opt,name=namespace,proto3" json:"namespace,omitempty"`
}

func (x *SourceRef) Reset() {
	*x = SourceRef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SourceRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SourceRef) ProtoMessage() {}

func (x *SourceRef) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SourceRef.ProtoReflect.Descriptor instead.
func (*SourceRef) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{0}
}

func (x *SourceRef) GetApiVersion() string {
	if x != nil {
		return x.ApiVersion
	}
	return ""
}

func (x *SourceRef) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *SourceRef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SourceRef) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

type Interval struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hours   int64 `protobuf:"varint,1,opt,name=hours,proto3" json:"hours,omitempty"`
	Minutes int64 `protobuf:"varint,2,opt,name=minutes,proto3" json:"minutes,omitempty"`
	Seconds int64 `protobuf:"varint,3,opt,name=seconds,proto3" json:"seconds,omitempty"`
}

func (x *Interval) Reset() {
	*x = Interval{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interval) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interval) ProtoMessage() {}

func (x *Interval) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interval.ProtoReflect.Descriptor instead.
func (*Interval) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{1}
}

func (x *Interval) GetHours() int64 {
	if x != nil {
		return x.Hours
	}
	return 0
}

func (x *Interval) GetMinutes() int64 {
	if x != nil {
		return x.Minutes
	}
	return 0
}

func (x *Interval) GetSeconds() int64 {
	if x != nil {
		return x.Seconds
	}
	return 0
}

type ResourceRef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type       string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Identifier string `protobuf:"bytes,3,opt,name=identifier,proto3" json:"identifier,omitempty"`
}

func (x *ResourceRef) Reset() {
	*x = ResourceRef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceRef) ProtoMessage() {}

func (x *ResourceRef) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceRef.ProtoReflect.Descriptor instead.
func (*ResourceRef) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{2}
}

func (x *ResourceRef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ResourceRef) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *ResourceRef) GetIdentifier() string {
	if x != nil {
		return x.Identifier
	}
	return ""
}

type NamespacedObjectReference struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
}

func (x *NamespacedObjectReference) Reset() {
	*x = NamespacedObjectReference{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NamespacedObjectReference) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NamespacedObjectReference) ProtoMessage() {}

func (x *NamespacedObjectReference) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NamespacedObjectReference.ProtoReflect.Descriptor instead.
func (*NamespacedObjectReference) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{3}
}

func (x *NamespacedObjectReference) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *NamespacedObjectReference) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

type ObjectRef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClusterName string `protobuf:"bytes,1,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Namespace   string `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Kind        string `protobuf:"bytes,4,opt,name=kind,proto3" json:"kind,omitempty"`
}

func (x *ObjectRef) Reset() {
	*x = ObjectRef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObjectRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObjectRef) ProtoMessage() {}

func (x *ObjectRef) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjectRef.ProtoReflect.Descriptor instead.
func (*ObjectRef) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{4}
}

func (x *ObjectRef) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *ObjectRef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ObjectRef) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ObjectRef) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

type TerraformObject struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                 string                       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace            string                       `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	ClusterName          string                       `protobuf:"bytes,3,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	Type                 string                       `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	Uid                  string                       `protobuf:"bytes,5,opt,name=uid,proto3" json:"uid,omitempty"`
	SourceRef            *SourceRef                   `protobuf:"bytes,6,opt,name=source_ref,json=sourceRef,proto3" json:"source_ref,omitempty"`
	AppliedRevision      string                       `protobuf:"bytes,7,opt,name=applied_revision,json=appliedRevision,proto3" json:"applied_revision,omitempty"`
	Path                 string                       `protobuf:"bytes,8,opt,name=path,proto3" json:"path,omitempty"`
	Interval             *Interval                    `protobuf:"bytes,9,opt,name=interval,proto3" json:"interval,omitempty"`
	LastUpdatedAt        string                       `protobuf:"bytes,10,opt,name=last_updated_at,json=lastUpdatedAt,proto3" json:"last_updated_at,omitempty"`
	DriftDetectionResult bool                         `protobuf:"varint,11,opt,name=drift_detection_result,json=driftDetectionResult,proto3" json:"drift_detection_result,omitempty"`
	Inventory            []*ResourceRef               `protobuf:"bytes,12,rep,name=inventory,proto3" json:"inventory,omitempty"`
	Conditions           []*Condition                 `protobuf:"bytes,13,rep,name=conditions,proto3" json:"conditions,omitempty"`
	Labels               map[string]string            `protobuf:"bytes,14,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Annotations          map[string]string            `protobuf:"bytes,15,rep,name=annotations,proto3" json:"annotations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	DependsOn            []*NamespacedObjectReference `protobuf:"bytes,16,rep,name=depends_on,json=dependsOn,proto3" json:"depends_on,omitempty"`
	Suspended            bool                         `protobuf:"varint,17,opt,name=suspended,proto3" json:"suspended,omitempty"`
}

func (x *TerraformObject) Reset() {
	*x = TerraformObject{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TerraformObject) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TerraformObject) ProtoMessage() {}

func (x *TerraformObject) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TerraformObject.ProtoReflect.Descriptor instead.
func (*TerraformObject) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{5}
}

func (x *TerraformObject) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TerraformObject) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *TerraformObject) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *TerraformObject) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *TerraformObject) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *TerraformObject) GetSourceRef() *SourceRef {
	if x != nil {
		return x.SourceRef
	}
	return nil
}

func (x *TerraformObject) GetAppliedRevision() string {
	if x != nil {
		return x.AppliedRevision
	}
	return ""
}

func (x *TerraformObject) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *TerraformObject) GetInterval() *Interval {
	if x != nil {
		return x.Interval
	}
	return nil
}

func (x *TerraformObject) GetLastUpdatedAt() string {
	if x != nil {
		return x.LastUpdatedAt
	}
	return ""
}

func (x *TerraformObject) GetDriftDetectionResult() bool {
	if x != nil {
		return x.DriftDetectionResult
	}
	return false
}

func (x *TerraformObject) GetInventory() []*ResourceRef {
	if x != nil {
		return x.Inventory
	}
	return nil
}

func (x *TerraformObject) GetConditions() []*Condition {
	if x != nil {
		return x.Conditions
	}
	return nil
}

func (x *TerraformObject) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *TerraformObject) GetAnnotations() map[string]string {
	if x != nil {
		return x.Annotations
	}
	return nil
}

func (x *TerraformObject) GetDependsOn() []*NamespacedObjectReference {
	if x != nil {
		return x.DependsOn
	}
	return nil
}

func (x *TerraformObject) GetSuspended() bool {
	if x != nil {
		return x.Suspended
	}
	return false
}

type Pagination struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PageSize  int32  `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	PageToken string `protobuf:"bytes,2,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
}

func (x *Pagination) Reset() {
	*x = Pagination{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pagination) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pagination) ProtoMessage() {}

func (x *Pagination) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pagination.ProtoReflect.Descriptor instead.
func (*Pagination) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{6}
}

func (x *Pagination) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *Pagination) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

type TerraformListError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClusterName string `protobuf:"bytes,1,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	Namespace   string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Message     string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *TerraformListError) Reset() {
	*x = TerraformListError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TerraformListError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TerraformListError) ProtoMessage() {}

func (x *TerraformListError) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TerraformListError.ProtoReflect.Descriptor instead.
func (*TerraformListError) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{7}
}

func (x *TerraformListError) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *TerraformListError) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *TerraformListError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type Condition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Status    string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Reason    string `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Message   string `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp string `protobuf:"bytes,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Condition) Reset() {
	*x = Condition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_terraform_types_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Condition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Condition) ProtoMessage() {}

func (x *Condition) ProtoReflect() protoreflect.Message {
	mi := &file_api_terraform_types_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Condition.ProtoReflect.Descriptor instead.
func (*Condition) Descriptor() ([]byte, []int) {
	return file_api_terraform_types_proto_rawDescGZIP(), []int{8}
}

func (x *Condition) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Condition) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Condition) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *Condition) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Condition) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

var File_api_terraform_types_proto protoreflect.FileDescriptor

var file_api_terraform_types_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x74, 0x65, 0x72,
	0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x22, 0x72, 0x0a, 0x09, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x52, 0x65, 0x66, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x70, 0x69, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x70, 0x69,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x54, 0x0a,
	0x08, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x68, 0x6f, 0x75,
	0x72, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x68, 0x6f, 0x75, 0x72, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x6d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x73, 0x65, 0x63, 0x6f,
	0x6e, 0x64, 0x73, 0x22, 0x55, 0x0a, 0x0b, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52,
	0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x22, 0x4d, 0x0a, 0x19, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x74, 0x0a, 0x09, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x66, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b,
	0x69, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x22,
	0xfd, 0x06, 0x0a, 0x0f, 0x54, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x75, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x36,
	0x0a, 0x0a, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x72, 0x65, 0x66, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x66, 0x52, 0x09, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x52, 0x65, 0x66, 0x12, 0x29, 0x0a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x65,
	0x64, 0x5f, 0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x65, 0x64, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x32, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61,
	0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66,
	0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x52,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x26, 0x0a, 0x0f, 0x6c, 0x61, 0x73,
	0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41,
	0x74, 0x12, 0x34, 0x0a, 0x16, 0x64, 0x72, 0x69, 0x66, 0x74, 0x5f, 0x64, 0x65, 0x74, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x14, 0x64, 0x72, 0x69, 0x66, 0x74, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x37, 0x0a, 0x09, 0x69, 0x6e, 0x76, 0x65, 0x6e,
	0x74, 0x6f, 0x72, 0x79, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x65, 0x72,
	0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x52, 0x65, 0x66, 0x52, 0x09, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
	0x12, 0x37, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0d,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d,
	0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x63,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x41, 0x0a, 0x06, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x74, 0x65, 0x72, 0x72,
	0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f,
	0x72, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x50, 0x0a, 0x0b,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x2e, 0x2e, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x2e, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x46,
	0x0a, 0x0a, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x73, 0x5f, 0x6f, 0x6e, 0x18, 0x10, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x27, 0x2e, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76,
	0x31, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x09, 0x64, 0x65, 0x70,
	0x65, 0x6e, 0x64, 0x73, 0x4f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x75, 0x73, 0x70, 0x65, 0x6e,
	0x64, 0x65, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x73, 0x75, 0x73, 0x70, 0x65,
	0x6e, 0x64, 0x65, 0x64, 0x1a, 0x39, 0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x3e, 0x0a, 0x10, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x48, 0x0a, 0x0a, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a,
	0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61,
	0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x6f, 0x0a, 0x12, 0x54, 0x65, 0x72,
	0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x21, 0x0a, 0x0c, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x87, 0x01, 0x0a, 0x09, 0x43,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x61, 0x76, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2f, 0x77, 0x65,
	0x61, 0x76, 0x65, 0x2d, 0x67, 0x69, 0x74, 0x6f, 0x70, 0x73, 0x2d, 0x65, 0x6e, 0x74, 0x65, 0x72,
	0x70, 0x72, 0x69, 0x73, 0x65, 0x2f, 0x74, 0x65, 0x72, 0x72, 0x61, 0x66, 0x6f, 0x72, 0x6d, 0x2f,
	0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_terraform_types_proto_rawDescOnce sync.Once
	file_api_terraform_types_proto_rawDescData = file_api_terraform_types_proto_rawDesc
)

func file_api_terraform_types_proto_rawDescGZIP() []byte {
	file_api_terraform_types_proto_rawDescOnce.Do(func() {
		file_api_terraform_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_terraform_types_proto_rawDescData)
	})
	return file_api_terraform_types_proto_rawDescData
}

var file_api_terraform_types_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_api_terraform_types_proto_goTypes = []interface{}{
	(*SourceRef)(nil),                 // 0: terraform.v1.SourceRef
	(*Interval)(nil),                  // 1: terraform.v1.Interval
	(*ResourceRef)(nil),               // 2: terraform.v1.ResourceRef
	(*NamespacedObjectReference)(nil), // 3: terraform.v1.NamespacedObjectReference
	(*ObjectRef)(nil),                 // 4: terraform.v1.ObjectRef
	(*TerraformObject)(nil),           // 5: terraform.v1.TerraformObject
	(*Pagination)(nil),                // 6: terraform.v1.Pagination
	(*TerraformListError)(nil),        // 7: terraform.v1.TerraformListError
	(*Condition)(nil),                 // 8: terraform.v1.Condition
	nil,                               // 9: terraform.v1.TerraformObject.LabelsEntry
	nil,                               // 10: terraform.v1.TerraformObject.AnnotationsEntry
}
var file_api_terraform_types_proto_depIdxs = []int32{
	0,  // 0: terraform.v1.TerraformObject.source_ref:type_name -> terraform.v1.SourceRef
	1,  // 1: terraform.v1.TerraformObject.interval:type_name -> terraform.v1.Interval
	2,  // 2: terraform.v1.TerraformObject.inventory:type_name -> terraform.v1.ResourceRef
	8,  // 3: terraform.v1.TerraformObject.conditions:type_name -> terraform.v1.Condition
	9,  // 4: terraform.v1.TerraformObject.labels:type_name -> terraform.v1.TerraformObject.LabelsEntry
	10, // 5: terraform.v1.TerraformObject.annotations:type_name -> terraform.v1.TerraformObject.AnnotationsEntry
	3,  // 6: terraform.v1.TerraformObject.depends_on:type_name -> terraform.v1.NamespacedObjectReference
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_api_terraform_types_proto_init() }
func file_api_terraform_types_proto_init() {
	if File_api_terraform_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_terraform_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SourceRef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interval); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceRef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NamespacedObjectReference); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjectRef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TerraformObject); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Pagination); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TerraformListError); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_terraform_types_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Condition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_terraform_types_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_terraform_types_proto_goTypes,
		DependencyIndexes: file_api_terraform_types_proto_depIdxs,
		MessageInfos:      file_api_terraform_types_proto_msgTypes,
	}.Build()
	File_api_terraform_types_proto = out.File
	file_api_terraform_types_proto_rawDesc = nil
	file_api_terraform_types_proto_goTypes = nil
	file_api_terraform_types_proto_depIdxs = nil
}
