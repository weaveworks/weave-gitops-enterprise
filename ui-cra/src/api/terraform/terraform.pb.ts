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
}

export class Terraform {
  static ListTerraformObjects(req: ListTerraformObjectsRequest, initReq?: fm.InitReq): Promise<ListTerraformObjectsResponse> {
    return fm.fetchReq<ListTerraformObjectsRequest, ListTerraformObjectsResponse>(`/v1/terraform_objects?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetTerraformObject(req: GetTerraformObjectRequest, initReq?: fm.InitReq): Promise<GetTerraformObjectResponse> {
    return fm.fetchReq<GetTerraformObjectRequest, GetTerraformObjectResponse>(`/v1/terraform_objects/${req["name"]}?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"})
  }
}