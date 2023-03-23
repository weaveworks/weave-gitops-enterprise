/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
export type QueryRequest = {
  query?: QueryClause[]
  offset?: number
  limit?: number
}

export type QueryClause = {
  key?: string
  value?: string
  operand?: string
}

export type QueryResponse = {
  objects?: Object[]
}

export type Object = {
  cluster?: string
  namespace?: string
  apiGroup?: string
  apiVersion?: string
  kind?: string
  name?: string
  status?: string
  message?: string
}

export type DebugGetAccessRulesRequest = {
}

export type DebugGetAccessRulesResponse = {
  rules?: AccessRule[]
}

export type AccessRule = {
  cluster?: string
  principal?: string
  namespace?: string
  accessibleKinds?: string[]
  subjects?: Subject[]
}

export type Role = {
  cluster?: string
  namespace?: string
  kind?: string
  name?: string
  policyRules?: PolicyRule[]
}

export type PolicyRule = {
  apiGroups?: string
  resources?: string
  verbs?: string
  roleId?: string
}

export type RoleBinding = {
  cluster?: string
  namespace?: string
  kind?: string
  name?: string
  roleRefName?: string
  roleRefKind?: string
  subjects?: Subject[]
}

export type Subject = {
  kind?: string
  name?: string
  namespace?: string
  apiGroup?: string
  roleBindingId?: string
}

export type StoreRolesRequest = {
  roles?: Role[]
}

export type StoreRolesResponse = {
}

export type StoreRoleBindingsRequest = {
  rolebindings?: RoleBinding[]
}

export type StoreRoleBindingsResponse = {
}

export type StoreObjectsRequest = {
  objects?: Object[]
}

export type StoreObjectsResponse = {
}

export type DeleteObjectsRequest = {
  objects?: Object[]
}

export type DeleteObjectsResponse = {
}

export type DeleteRolesRequest = {
  roles?: Role[]
}

export type DeleteRolesResponse = {
}

export type DeleteRoleBindingsRequest = {
  rolebindings?: RoleBinding[]
}

export type DeleteRoleBindingsResponse = {
}

export class Query {
  static DoQuery(req: QueryRequest, initReq?: fm.InitReq): Promise<QueryResponse> {
    return fm.fetchReq<QueryRequest, QueryResponse>(`/v1/query`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static DebugGetAccessRules(req: DebugGetAccessRulesRequest, initReq?: fm.InitReq): Promise<DebugGetAccessRulesResponse> {
    return fm.fetchReq<DebugGetAccessRulesRequest, DebugGetAccessRulesResponse>(`/v1/debug/access-rules?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static StoreRoles(req: StoreRolesRequest, initReq?: fm.InitReq): Promise<StoreRolesResponse> {
    return fm.fetchReq<StoreRolesRequest, StoreRolesResponse>(`/v1/query/roles`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static StoreRoleBindings(req: StoreRoleBindingsRequest, initReq?: fm.InitReq): Promise<StoreRoleBindingsResponse> {
    return fm.fetchReq<StoreRoleBindingsRequest, StoreRoleBindingsResponse>(`/v1/query/rolebindings`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static StoreObjects(req: StoreObjectsRequest, initReq?: fm.InitReq): Promise<StoreObjectsResponse> {
    return fm.fetchReq<StoreObjectsRequest, StoreObjectsResponse>(`/v1/query/objects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static DeleteObjects(req: DeleteObjectsRequest, initReq?: fm.InitReq): Promise<DeleteObjectsResponse> {
    return fm.fetchReq<DeleteObjectsRequest, DeleteObjectsResponse>(`/v1/query/objects`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
  static DeleteeRoles(req: DeleteRolesRequest, initReq?: fm.InitReq): Promise<DeleteRolesResponse> {
    return fm.fetchReq<DeleteRolesRequest, DeleteRolesResponse>(`/v1/query/roles`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
  static DeleteRoleBindings(req: DeleteRoleBindingsRequest, initReq?: fm.InitReq): Promise<DeleteRoleBindingsResponse> {
    return fm.fetchReq<DeleteRoleBindingsRequest, DeleteRoleBindingsResponse>(`/v1/query/rolebindings`, {...initReq, method: "DELETE", body: JSON.stringify(req)})
  }
}