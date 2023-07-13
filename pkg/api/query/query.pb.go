//
// This file holds the protobuf definitions
// for the Weave GitOps Query Service API.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: api/query/query.proto

package api

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type QueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Terms     string   `protobuf:"bytes,1,opt,name=terms,proto3" json:"terms,omitempty"`
	Filters   []string `protobuf:"bytes,2,rep,name=filters,proto3" json:"filters,omitempty"`
	Offset    int32    `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit     int32    `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	OrderBy   string   `protobuf:"bytes,5,opt,name=orderBy,proto3" json:"orderBy,omitempty"`
	Ascending bool     `protobuf:"varint,6,opt,name=ascending,proto3" json:"ascending,omitempty"`
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{0}
}

func (x *QueryRequest) GetTerms() string {
	if x != nil {
		return x.Terms
	}
	return ""
}

func (x *QueryRequest) GetFilters() []string {
	if x != nil {
		return x.Filters
	}
	return nil
}

func (x *QueryRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *QueryRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *QueryRequest) GetOrderBy() string {
	if x != nil {
		return x.OrderBy
	}
	return ""
}

func (x *QueryRequest) GetAscending() bool {
	if x != nil {
		return x.Ascending
	}
	return false
}

type QueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Objects []*Object `protobuf:"bytes,1,rep,name=objects,proto3" json:"objects,omitempty"`
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{1}
}

func (x *QueryResponse) GetObjects() []*Object {
	if x != nil {
		return x.Objects
	}
	return nil
}

type Object struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cluster      string `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Namespace    string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Kind         string `protobuf:"bytes,3,opt,name=kind,proto3" json:"kind,omitempty"`
	Name         string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Status       string `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
	ApiGroup     string `protobuf:"bytes,6,opt,name=apiGroup,proto3" json:"apiGroup,omitempty"`
	ApiVersion   string `protobuf:"bytes,7,opt,name=apiVersion,proto3" json:"apiVersion,omitempty"`
	Message      string `protobuf:"bytes,8,opt,name=message,proto3" json:"message,omitempty"`
	Category     string `protobuf:"bytes,9,opt,name=category,proto3" json:"category,omitempty"`
	Unstructured string `protobuf:"bytes,10,opt,name=unstructured,proto3" json:"unstructured,omitempty"`
}

func (x *Object) Reset() {
	*x = Object{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Object) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Object) ProtoMessage() {}

func (x *Object) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Object.ProtoReflect.Descriptor instead.
func (*Object) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{2}
}

func (x *Object) GetCluster() string {
	if x != nil {
		return x.Cluster
	}
	return ""
}

func (x *Object) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *Object) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Object) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Object) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Object) GetApiGroup() string {
	if x != nil {
		return x.ApiGroup
	}
	return ""
}

func (x *Object) GetApiVersion() string {
	if x != nil {
		return x.ApiVersion
	}
	return ""
}

func (x *Object) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Object) GetCategory() string {
	if x != nil {
		return x.Category
	}
	return ""
}

func (x *Object) GetUnstructured() string {
	if x != nil {
		return x.Unstructured
	}
	return ""
}

type DebugGetAccessRulesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DebugGetAccessRulesRequest) Reset() {
	*x = DebugGetAccessRulesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DebugGetAccessRulesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DebugGetAccessRulesRequest) ProtoMessage() {}

func (x *DebugGetAccessRulesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DebugGetAccessRulesRequest.ProtoReflect.Descriptor instead.
func (*DebugGetAccessRulesRequest) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{3}
}

type DebugGetAccessRulesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rules []*AccessRule `protobuf:"bytes,1,rep,name=rules,proto3" json:"rules,omitempty"`
}

func (x *DebugGetAccessRulesResponse) Reset() {
	*x = DebugGetAccessRulesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DebugGetAccessRulesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DebugGetAccessRulesResponse) ProtoMessage() {}

func (x *DebugGetAccessRulesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DebugGetAccessRulesResponse.ProtoReflect.Descriptor instead.
func (*DebugGetAccessRulesResponse) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{4}
}

func (x *DebugGetAccessRulesResponse) GetRules() []*AccessRule {
	if x != nil {
		return x.Rules
	}
	return nil
}

type AccessRule struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cluster           string     `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
	Principal         string     `protobuf:"bytes,2,opt,name=principal,proto3" json:"principal,omitempty"`
	Namespace         string     `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	AccessibleKinds   []string   `protobuf:"bytes,4,rep,name=accessibleKinds,proto3" json:"accessibleKinds,omitempty"`
	Subjects          []*Subject `protobuf:"bytes,5,rep,name=subjects,proto3" json:"subjects,omitempty"`
	ProvidedByRole    string     `protobuf:"bytes,6,opt,name=providedByRole,proto3" json:"providedByRole,omitempty"`
	ProvidedByBinding string     `protobuf:"bytes,7,opt,name=providedByBinding,proto3" json:"providedByBinding,omitempty"`
}

func (x *AccessRule) Reset() {
	*x = AccessRule{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccessRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccessRule) ProtoMessage() {}

func (x *AccessRule) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccessRule.ProtoReflect.Descriptor instead.
func (*AccessRule) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{5}
}

func (x *AccessRule) GetCluster() string {
	if x != nil {
		return x.Cluster
	}
	return ""
}

func (x *AccessRule) GetPrincipal() string {
	if x != nil {
		return x.Principal
	}
	return ""
}

func (x *AccessRule) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *AccessRule) GetAccessibleKinds() []string {
	if x != nil {
		return x.AccessibleKinds
	}
	return nil
}

func (x *AccessRule) GetSubjects() []*Subject {
	if x != nil {
		return x.Subjects
	}
	return nil
}

func (x *AccessRule) GetProvidedByRole() string {
	if x != nil {
		return x.ProvidedByRole
	}
	return ""
}

func (x *AccessRule) GetProvidedByBinding() string {
	if x != nil {
		return x.ProvidedByBinding
	}
	return ""
}

type Subject struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kind string `protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Subject) Reset() {
	*x = Subject{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Subject) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Subject) ProtoMessage() {}

func (x *Subject) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Subject.ProtoReflect.Descriptor instead.
func (*Subject) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{6}
}

func (x *Subject) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Subject) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ListFacetsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListFacetsRequest) Reset() {
	*x = ListFacetsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFacetsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFacetsRequest) ProtoMessage() {}

func (x *ListFacetsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFacetsRequest.ProtoReflect.Descriptor instead.
func (*ListFacetsRequest) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{7}
}

type ListFacetsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Facets []*Facet `protobuf:"bytes,1,rep,name=facets,proto3" json:"facets,omitempty"`
}

func (x *ListFacetsResponse) Reset() {
	*x = ListFacetsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFacetsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFacetsResponse) ProtoMessage() {}

func (x *ListFacetsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFacetsResponse.ProtoReflect.Descriptor instead.
func (*ListFacetsResponse) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{8}
}

func (x *ListFacetsResponse) GetFacets() []*Facet {
	if x != nil {
		return x.Facets
	}
	return nil
}

type Facet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Field  string   `protobuf:"bytes,1,opt,name=field,proto3" json:"field,omitempty"`
	Values []string `protobuf:"bytes,2,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *Facet) Reset() {
	*x = Facet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_query_query_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Facet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Facet) ProtoMessage() {}

func (x *Facet) ProtoReflect() protoreflect.Message {
	mi := &file_api_query_query_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Facet.ProtoReflect.Descriptor instead.
func (*Facet) Descriptor() ([]byte, []int) {
	return file_api_query_query_proto_rawDescGZIP(), []int{9}
}

func (x *Facet) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *Facet) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_api_query_query_proto protoreflect.FileDescriptor

var file_api_query_query_proto_rawDesc = []byte{
	0x0a, 0x15, 0x61, 0x70, 0x69, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xa4, 0x01, 0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x65, 0x72, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x65, 0x72, 0x6d, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73,
	0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69,
	0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x73, 0x63, 0x65,
	0x6e, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x61, 0x73, 0x63,
	0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x22, 0x3b, 0x0a, 0x0d, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x07, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x07, 0x6f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x73, 0x22, 0x96, 0x02, 0x0a, 0x06, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d,
	0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x70, 0x69, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x70, 0x69, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x70, 0x69, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x70, 0x69, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x75, 0x6e, 0x73, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x65, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x75, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x65, 0x64, 0x22, 0x1c, 0x0a, 0x1a,
	0x44, 0x65, 0x62, 0x75, 0x67, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75,
	0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x49, 0x0a, 0x1b, 0x44, 0x65,
	0x62, 0x75, 0x67, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x05, 0x72, 0x75, 0x6c,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x05,
	0x72, 0x75, 0x6c, 0x65, 0x73, 0x22, 0x91, 0x02, 0x0a, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x52, 0x75, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x1c,
	0x0a, 0x09, 0x70, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x70, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x12, 0x1c, 0x0a, 0x09,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x61, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x69, 0x62, 0x6c, 0x65, 0x4b, 0x69, 0x6e, 0x64, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x0f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x69, 0x62, 0x6c, 0x65, 0x4b,
	0x69, 0x6e, 0x64, 0x73, 0x12, 0x2d, 0x0a, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x64, 0x42,
	0x79, 0x52, 0x6f, 0x6c, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x64, 0x42, 0x79, 0x52, 0x6f, 0x6c, 0x65, 0x12, 0x2c, 0x0a, 0x11, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x64, 0x42, 0x79, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x64,
	0x42, 0x79, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x22, 0x31, 0x0a, 0x07, 0x53, 0x75, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x13, 0x0a, 0x11,
	0x4c, 0x69, 0x73, 0x74, 0x46, 0x61, 0x63, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x3d, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x61, 0x63, 0x65, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x66, 0x61, 0x63, 0x65, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x61, 0x63, 0x65, 0x74, 0x52, 0x06, 0x66, 0x61, 0x63, 0x65, 0x74, 0x73,
	0x22, 0x35, 0x0a, 0x05, 0x46, 0x61, 0x63, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x32, 0xbb, 0x02, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x12, 0x50, 0x0a, 0x07, 0x44, 0x6f, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x16, 0x2e, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x3a, 0x01, 0x2a, 0x22, 0x09, 0x2f, 0x76, 0x31, 0x2f, 0x71, 0x75,
	0x65, 0x72, 0x79, 0x12, 0x5b, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x61, 0x63, 0x65, 0x74,
	0x73, 0x12, 0x1b, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x46, 0x61, 0x63, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c,
	0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x61,
	0x63, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x12, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x61, 0x63, 0x65, 0x74, 0x73,
	0x12, 0x82, 0x01, 0x0a, 0x13, 0x44, 0x65, 0x62, 0x75, 0x67, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x12, 0x24, 0x2e, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x62, 0x75, 0x67, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25,
	0x2e, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x62, 0x75, 0x67, 0x47,
	0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x18, 0x12, 0x16, 0x2f,
	0x76, 0x31, 0x2f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x2d,
	0x72, 0x75, 0x6c, 0x65, 0x73, 0x42, 0xd3, 0x01, 0x92, 0x41, 0x96, 0x01, 0x12, 0x70, 0x0a, 0x1e,
	0x57, 0x65, 0x61, 0x76, 0x65, 0x20, 0x47, 0x69, 0x74, 0x4f, 0x70, 0x73, 0x20, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x20, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x20, 0x41, 0x50, 0x49, 0x12, 0x49,
	0x54, 0x68, 0x65, 0x20, 0x41, 0x50, 0x49, 0x20, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x20,
	0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x20, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x2d, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x20, 0x71, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x20, 0x66, 0x6f,
	0x72, 0x20, 0x57, 0x65, 0x61, 0x76, 0x65, 0x20, 0x47, 0x69, 0x74, 0x4f, 0x70, 0x73, 0x20, 0x45,
	0x6e, 0x74, 0x65, 0x72, 0x70, 0x72, 0x69, 0x73, 0x65, 0x32, 0x03, 0x30, 0x2e, 0x31, 0x32, 0x10,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e,
	0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73,
	0x6f, 0x6e, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77,
	0x65, 0x61, 0x76, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2f, 0x77, 0x65, 0x61, 0x76, 0x65, 0x2d,
	0x67, 0x69, 0x74, 0x6f, 0x70, 0x73, 0x2d, 0x65, 0x6e, 0x74, 0x65, 0x72, 0x70, 0x72, 0x69, 0x73,
	0x65, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_api_query_query_proto_rawDescOnce sync.Once
	file_api_query_query_proto_rawDescData = file_api_query_query_proto_rawDesc
)

func file_api_query_query_proto_rawDescGZIP() []byte {
	file_api_query_query_proto_rawDescOnce.Do(func() {
		file_api_query_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_query_query_proto_rawDescData)
	})
	return file_api_query_query_proto_rawDescData
}

var file_api_query_query_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_api_query_query_proto_goTypes = []interface{}{
	(*QueryRequest)(nil),                // 0: query.v1.QueryRequest
	(*QueryResponse)(nil),               // 1: query.v1.QueryResponse
	(*Object)(nil),                      // 2: query.v1.Object
	(*DebugGetAccessRulesRequest)(nil),  // 3: query.v1.DebugGetAccessRulesRequest
	(*DebugGetAccessRulesResponse)(nil), // 4: query.v1.DebugGetAccessRulesResponse
	(*AccessRule)(nil),                  // 5: query.v1.AccessRule
	(*Subject)(nil),                     // 6: query.v1.Subject
	(*ListFacetsRequest)(nil),           // 7: query.v1.ListFacetsRequest
	(*ListFacetsResponse)(nil),          // 8: query.v1.ListFacetsResponse
	(*Facet)(nil),                       // 9: query.v1.Facet
}
var file_api_query_query_proto_depIdxs = []int32{
	2, // 0: query.v1.QueryResponse.objects:type_name -> query.v1.Object
	5, // 1: query.v1.DebugGetAccessRulesResponse.rules:type_name -> query.v1.AccessRule
	6, // 2: query.v1.AccessRule.subjects:type_name -> query.v1.Subject
	9, // 3: query.v1.ListFacetsResponse.facets:type_name -> query.v1.Facet
	0, // 4: query.v1.Query.DoQuery:input_type -> query.v1.QueryRequest
	7, // 5: query.v1.Query.ListFacets:input_type -> query.v1.ListFacetsRequest
	3, // 6: query.v1.Query.DebugGetAccessRules:input_type -> query.v1.DebugGetAccessRulesRequest
	1, // 7: query.v1.Query.DoQuery:output_type -> query.v1.QueryResponse
	8, // 8: query.v1.Query.ListFacets:output_type -> query.v1.ListFacetsResponse
	4, // 9: query.v1.Query.DebugGetAccessRules:output_type -> query.v1.DebugGetAccessRulesResponse
	7, // [7:10] is the sub-list for method output_type
	4, // [4:7] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_api_query_query_proto_init() }
func file_api_query_query_proto_init() {
	if File_api_query_query_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_query_query_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryRequest); i {
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
		file_api_query_query_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryResponse); i {
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
		file_api_query_query_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Object); i {
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
		file_api_query_query_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DebugGetAccessRulesRequest); i {
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
		file_api_query_query_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DebugGetAccessRulesResponse); i {
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
		file_api_query_query_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccessRule); i {
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
		file_api_query_query_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Subject); i {
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
		file_api_query_query_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFacetsRequest); i {
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
		file_api_query_query_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFacetsResponse); i {
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
		file_api_query_query_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Facet); i {
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
			RawDescriptor: file_api_query_query_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_query_query_proto_goTypes,
		DependencyIndexes: file_api_query_query_proto_depIdxs,
		MessageInfos:      file_api_query_query_proto_msgTypes,
	}.Build()
	File_api_query_query_proto = out.File
	file_api_query_query_proto_rawDesc = nil
	file_api_query_query_proto_goTypes = nil
	file_api_query_query_proto_depIdxs = nil
}
