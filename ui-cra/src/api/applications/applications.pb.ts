/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"

export enum GitProvider {
  Unknown = "Unknown",
  GitHub = "GitHub",
  GitLab = "GitLab",
}

export type AuthenticateRequest = {
  providerName?: string
  accessToken?: string
}

export type AuthenticateResponse = {
  token?: string
}

export type GetGithubDeviceCodeRequest = {
}

export type GetGithubDeviceCodeResponse = {
  userCode?: string
  deviceCode?: string
  validationURI?: string
  interval?: number
}

export type GetGithubAuthStatusRequest = {
  deviceCode?: string
}

export type GetGithubAuthStatusResponse = {
  accessToken?: string
  error?: string
}

export type ParseRepoURLRequest = {
  url?: string
}

export type ParseRepoURLResponse = {
  name?: string
  provider?: GitProvider
  owner?: string
}

export type GetGitlabAuthURLRequest = {
  redirectUri?: string
}

export type GetGitlabAuthURLResponse = {
  url?: string
}

export type AuthorizeGitlabRequest = {
  code?: string
  redirectUri?: string
}

export type AuthorizeGitlabResponse = {
  token?: string
}

export type ValidateProviderTokenRequest = {
  provider?: GitProvider
}

export type ValidateProviderTokenResponse = {
  valid?: boolean
}

export class Applications {
  static Authenticate(req: AuthenticateRequest, initReq?: fm.InitReq): Promise<AuthenticateResponse> {
    return fm.fetchReq<AuthenticateRequest, AuthenticateResponse>(`/v1/authenticate/${req["providerName"]}`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetGithubDeviceCode(req: GetGithubDeviceCodeRequest, initReq?: fm.InitReq): Promise<GetGithubDeviceCodeResponse> {
    return fm.fetchReq<GetGithubDeviceCodeRequest, GetGithubDeviceCodeResponse>(`/v1/applications/auth_providers/github?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetGithubAuthStatus(req: GetGithubAuthStatusRequest, initReq?: fm.InitReq): Promise<GetGithubAuthStatusResponse> {
    return fm.fetchReq<GetGithubAuthStatusRequest, GetGithubAuthStatusResponse>(`/v1/applications/auth_providers/github/status`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetGitlabAuthURL(req: GetGitlabAuthURLRequest, initReq?: fm.InitReq): Promise<GetGitlabAuthURLResponse> {
    return fm.fetchReq<GetGitlabAuthURLRequest, GetGitlabAuthURLResponse>(`/v1/applications/auth_providers/gitlab?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static AuthorizeGitlab(req: AuthorizeGitlabRequest, initReq?: fm.InitReq): Promise<AuthorizeGitlabResponse> {
    return fm.fetchReq<AuthorizeGitlabRequest, AuthorizeGitlabResponse>(`/v1/applications/auth_providers/gitlab/authorize`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ParseRepoURL(req: ParseRepoURLRequest, initReq?: fm.InitReq): Promise<ParseRepoURLResponse> {
    return fm.fetchReq<ParseRepoURLRequest, ParseRepoURLResponse>(`/v1/applications/parse_repo_url?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ValidateProviderToken(req: ValidateProviderTokenRequest, initReq?: fm.InitReq): Promise<ValidateProviderTokenResponse> {
    return fm.fetchReq<ValidateProviderTokenRequest, ValidateProviderTokenResponse>(`/v1/applications/validate_token`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}