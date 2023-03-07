/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
export type QueryRequest = {
  query?: QueryOpts
}

export type QueryOpts = {
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
  kind?: string
  name?: string
  status?: string
}

export class Query {
  static DoQuery(req: QueryRequest, initReq?: fm.InitReq): Promise<QueryResponse> {
    return fm.fetchReq<QueryRequest, QueryResponse>(`/v1/query`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}