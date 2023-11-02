/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"

export enum EnabledComponent {
  unknown = "unknown",
  applications = "applications",
  sources = "sources",
  gitopssets = "gitopssets",
  templates = "templates",
}

export type QueryRequest = {
  terms?: string
  filters?: string[]
  offset?: number
  limit?: number
  orderBy?: string
  ascending?: boolean
}

export type QueryResponse = {
  objects?: Object[]
}

export type Object = {
  cluster?: string
  namespace?: string
  kind?: string
  name?: string
  status?: string
  apiGroup?: string
  apiVersion?: string
  message?: string
  category?: string
  unstructured?: string
  id?: string
  tenant?: string
  labels?: {[key: string]: string}
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
  providedByRole?: string
  providedByBinding?: string
}

export type Subject = {
  kind?: string
  name?: string
}

export type ListFacetsRequest = {
}

export type ListFacetsResponse = {
  facets?: Facet[]
}

export type Facet = {
  field?: string
  values?: string[]
}

export type ListEnabledComponentsRequest = {
}

export type ListEnabledComponentsResponse = {
  components?: EnabledComponent[]
}

export class Query {
  static DoQuery(req: QueryRequest, initReq?: fm.InitReq): Promise<QueryResponse> {
    return fm.fetchReq<QueryRequest, QueryResponse>(`/v1/query`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListFacets(req: ListFacetsRequest, initReq?: fm.InitReq): Promise<ListFacetsResponse> {
    return fm.fetchReq<ListFacetsRequest, ListFacetsResponse>(`/v1/facets?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static DebugGetAccessRules(req: DebugGetAccessRulesRequest, initReq?: fm.InitReq): Promise<DebugGetAccessRulesResponse> {
    return fm.fetchReq<DebugGetAccessRulesRequest, DebugGetAccessRulesResponse>(`/v1/debug/access-rules?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ListEnabledComponents(req: ListEnabledComponentsRequest, initReq?: fm.InitReq): Promise<ListEnabledComponentsResponse> {
    return fm.fetchReq<ListEnabledComponentsRequest, ListEnabledComponentsResponse>(`/v1/enabled-components?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}