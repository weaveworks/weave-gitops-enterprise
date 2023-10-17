/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufDuration from "../../google/protobuf/duration.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);
export type GetYAMLRequest = {
  type?: string
  resource?: string
}

export type GetYAMLResponse = {
  yaml?: string
}


type BaseGitRepository = {
  name?: string
  namespace?: string
  url?: string
  interval?: GoogleProtobufDuration.Duration
}

export type GitRepository = BaseGitRepository
  & OneOf<{ branch: string }>
  & OneOf<{ tag: string }>
  & OneOf<{ semver: string }>
  & OneOf<{ commit: string }>
  & OneOf<{ refName: string }>
  & OneOf<{ secretRefName: string }>


type BaseHelmRepository = {
  name?: string
  namespace?: string
  url?: string
  interval?: GoogleProtobufDuration.Duration
}

export type HelmRepository = BaseHelmRepository
  & OneOf<{ type: string }>
  & OneOf<{ provider: string }>
  & OneOf<{ secretRefName: string }>
  & OneOf<{ passCredentials: boolean }>


type BaseBucket = {
  name?: string
  namespace?: string
  bucketName?: string
  endpoint?: string
  interval?: GoogleProtobufDuration.Duration
}

export type Bucket = BaseBucket
  & OneOf<{ provider: string }>
  & OneOf<{ secretRefName: string }>
  & OneOf<{ region: string }>
  & OneOf<{ insecure: boolean }>


type BaseOCIRepository = {
  name?: string
  namespace?: string
  url?: string
  interval?: GoogleProtobufDuration.Duration
}

export type OCIRepository = BaseOCIRepository
  & OneOf<{ provider: string }>
  & OneOf<{ secretRefName: string }>
  & OneOf<{ serviceAccountName: string }>
  & OneOf<{ certSecretRefName: string }>
  & OneOf<{ insecure: boolean }>
  & OneOf<{ tag: string }>
  & OneOf<{ semver: string }>
  & OneOf<{ digest: string }>