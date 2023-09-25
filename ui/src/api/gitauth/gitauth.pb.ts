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
  BitBucketServer = "BitBucketServer",
  AzureDevOps = "AzureDevOps",
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

export type GetBitbucketServerAuthURLRequest = {
  redirectUri?: string
}

export type GetBitbucketServerAuthURLResponse = {
  url?: string
}

export type AuthorizeBitbucketServerRequest = {
  code?: string
  state?: string
  redirectUri?: string
}

export type AuthorizeBitbucketServerResponse = {
  token?: string
}

export type GetAzureDevOpsAuthURLRequest = {
  redirectUri?: string
}

export type GetAzureDevOpsAuthURLResponse = {
  url?: string
}

export type AuthorizeAzureDevOpsRequest = {
  code?: string
  state?: string
  redirectUri?: string
}

export type AuthorizeAzureDevOpsResponse = {
  token?: string
}

export class GitAuth {
  static Authenticate(req: AuthenticateRequest, initReq?: fm.InitReq): Promise<AuthenticateResponse> {
    return fm.fetchReq<AuthenticateRequest, AuthenticateResponse>(`/v1/authenticate/${req["providerName"]}`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetGithubDeviceCode(req: GetGithubDeviceCodeRequest, initReq?: fm.InitReq): Promise<GetGithubDeviceCodeResponse> {
    return fm.fetchReq<GetGithubDeviceCodeRequest, GetGithubDeviceCodeResponse>(`/v1/gitauth/auth-providers/github?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetGithubAuthStatus(req: GetGithubAuthStatusRequest, initReq?: fm.InitReq): Promise<GetGithubAuthStatusResponse> {
    return fm.fetchReq<GetGithubAuthStatusRequest, GetGithubAuthStatusResponse>(`/v1/gitauth/auth-providers/github/status`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetGitlabAuthURL(req: GetGitlabAuthURLRequest, initReq?: fm.InitReq): Promise<GetGitlabAuthURLResponse> {
    return fm.fetchReq<GetGitlabAuthURLRequest, GetGitlabAuthURLResponse>(`/v1/gitauth/auth-providers/gitlab?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetBitbucketServerAuthURL(req: GetBitbucketServerAuthURLRequest, initReq?: fm.InitReq): Promise<GetBitbucketServerAuthURLResponse> {
    return fm.fetchReq<GetBitbucketServerAuthURLRequest, GetBitbucketServerAuthURLResponse>(`/v1/gitauth/auth-providers/bitbucketserver?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static AuthorizeBitbucketServer(req: AuthorizeBitbucketServerRequest, initReq?: fm.InitReq): Promise<AuthorizeBitbucketServerResponse> {
    return fm.fetchReq<AuthorizeBitbucketServerRequest, AuthorizeBitbucketServerResponse>(`/v1/gitauth/auth-providers/bitbucketserver/authorize`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static AuthorizeGitlab(req: AuthorizeGitlabRequest, initReq?: fm.InitReq): Promise<AuthorizeGitlabResponse> {
    return fm.fetchReq<AuthorizeGitlabRequest, AuthorizeGitlabResponse>(`/v1/gitauth/auth-providers/gitlab/authorize`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetAzureDevOpsAuthURL(req: GetAzureDevOpsAuthURLRequest, initReq?: fm.InitReq): Promise<GetAzureDevOpsAuthURLResponse> {
    return fm.fetchReq<GetAzureDevOpsAuthURLRequest, GetAzureDevOpsAuthURLResponse>(`/v1/gitauth/auth-providers/azuredevops?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static AuthorizeAzureDevOps(req: AuthorizeAzureDevOpsRequest, initReq?: fm.InitReq): Promise<AuthorizeAzureDevOpsResponse> {
    return fm.fetchReq<AuthorizeAzureDevOpsRequest, AuthorizeAzureDevOpsResponse>(`/v1/gitauth/auth-providers/azuredevops/authorize`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ParseRepoURL(req: ParseRepoURLRequest, initReq?: fm.InitReq): Promise<ParseRepoURLResponse> {
    return fm.fetchReq<ParseRepoURLRequest, ParseRepoURLResponse>(`/v1/gitauth/parse-repo-url?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static ValidateProviderToken(req: ValidateProviderTokenRequest, initReq?: fm.InitReq): Promise<ValidateProviderTokenResponse> {
    return fm.fetchReq<ValidateProviderTokenRequest, ValidateProviderTokenResponse>(`/v1/gitauth/validate-token`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}