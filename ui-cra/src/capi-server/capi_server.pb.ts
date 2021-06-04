/* eslint-disable */
// @ts-nocheck
/*
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from './fetch.pb';
export type ListTemplatesRequest = {};

export type Param = {
  name?: string;
  description?: string;
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  params?: Param[];
};

export type ListTemplatesResponse = {
  templates?: Template[];
  total?: number;
};

export class ClustersService {
  static ListTemplates(
    req: ListTemplatesRequest,
    initReq?: fm.InitReq,
  ): Promise<ListTemplatesResponse> {
    return fm.fetchReq<ListTemplatesRequest, ListTemplatesResponse>(
      `/v1/templates?${fm.renderURLSearchParams(req, [])}`,
      { ...initReq, method: 'GET' },
    );
  }
}
