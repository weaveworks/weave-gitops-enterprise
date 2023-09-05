export type Unresolved = any;

/**
 * @example `"#fff"`
 * @example `"#FFFFFF"
 * @example `"#ffffffFF"
 */
export type HexColor = string;

/**
 * @example `hsl(100, 100%, 100%)
 */
export type HSLColor = string;

/**
 * @example `1566511955`
 */
export type UnixTimestampSeconds = number;

/**
 * @example `1566511955123`
 */
export type UnixTimestampMilliseconds = number;

export type RequestMethod =
  | 'GET'
  | 'HEAD'
  | 'POST'
  | 'PUT'
  | 'DELETE'
  | 'CONNECT'
  | 'OPTIONS'
  | 'TRACE';

export type URL = string;

/**
 * Git HTTP URLs come in the form `https://example.com/gitproject.git`.
 * @see https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols
 */
export type GitHttpUrl = URL;

/**
 * Git SSH URLs come in the form  `ssh://[user@]server/project.git` or `[user@]server:project.git`.
 * @see https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols
 */
export type GitSshUrl = string;

export type GitURL = GitHttpUrl | GitSshUrl;

export type GitBranchName = string;

export type HandlebarTemplate = string;

/**
 * Time intervals, e.g. `"5m"` representing 5 minutes.
 */
export type Duration = string;

/**
 * useful when extracting the type of the value of an object.
 */
export type ValueOf<T> = T[keyof T];

export type Seconds = number;

export type Milliseconds = number;
