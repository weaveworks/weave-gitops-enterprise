/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as TerraformV1Types from "./types.pb"
export type ListTerraformObjectsRequest = {
  clusterName?: string
  namespace?: string
  pagination?: TerraformV1Types.Pagination
}

export type ListTerraformObjectsResponse = {
  objects?: TerraformV1Types.TerraformObject[]
  errors?: TerraformV1Types.TerraformListError[]
}

export type GetTerraformObjectRequest = {
  clusterName?: string
  name?: string
  namespace?: string
}

export type GetTerraformObjectResponse = {
  object?: TerraformV1Types.TerraformObject
  yaml?: string
  type?: string
}

export type SyncTerraformObjectsRequest = {
  objects?: TerraformV1Types.ObjectRef[]
}

export type SyncTerraformObjectsResponse = {
  success?: boolean
}

export type ToggleSuspendTerraformObjectsRequest = {
  objects?: TerraformV1Types.ObjectRef[]
  suspend?: boolean
}

export type ToggleSuspendTerraformObjectsResponse = {
}

export type GetTerraformObjectPlanRequest = {
  clusterName?: string
  name?: string
  namespace?: string
}

export type GetTerraformObjectPlanResponse = {
  plan?: string
  enablePlanViewing?: boolean
  error?: string
}

export type ReplanTerraformObjectRequest = {
  clusterName?: string
  name?: string
  namespace?: string
}

export type ReplanTerraformObjectResponse = {
  replanRequested?: boolean
}

export class Terraform {
  static ListTerraformObjects(req: ListTerraformObjectsRequest, initReq?: fm.InitReq): Promise<ListTerraformObjectsResponse> {
    return fm.fetchReq<ListTerraformObjectsRequest, ListTerraformObjectsResponse>(`/v1/terraform_objects?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetTerraformObject(req: GetTerraformObjectRequest, initReq?: fm.InitReq): Promise<GetTerraformObjectResponse> {
    return fm.fetchReq<GetTerraformObjectRequest, GetTerraformObjectResponse>(`/v1/terraform_objects/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
  static SyncTerraformObjects(req: SyncTerraformObjectsRequest, initReq?: fm.InitReq): Promise<SyncTerraformObjectsResponse> {
    return fm.fetchReq<SyncTerraformObjectsRequest, SyncTerraformObjectsResponse>(`/v1/terraform_objects/sync`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ToggleSuspendTerraformObjects(req: ToggleSuspendTerraformObjectsRequest, initReq?: fm.InitReq): Promise<ToggleSuspendTerraformObjectsResponse> {
    return fm.fetchReq<ToggleSuspendTerraformObjectsRequest, ToggleSuspendTerraformObjectsResponse>(`/v1/terraform_objects/suspend`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetTerraformObjectPlan(req: GetTerraformObjectPlanRequest, initReq?: fm.InitReq): Promise<GetTerraformObjectPlanResponse> {
    return fm.fetchReq<GetTerraformObjectPlanRequest, GetTerraformObjectPlanResponse>(`/v1/terraform_objects/plan?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ReplanTerraformObject(req: ReplanTerraformObjectRequest, initReq?: fm.InitReq): Promise<ReplanTerraformObjectResponse> {
    return fm.fetchReq<ReplanTerraformObjectRequest, ReplanTerraformObjectResponse>(`/v1/terraform_objects/replan`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}