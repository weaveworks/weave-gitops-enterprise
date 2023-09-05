export type RequestMethod =
  | 'GET'
  | 'HEAD'
  | 'POST'
  | 'PUT'
  | 'DELETE'
  | 'CONNECT'
  | 'OPTIONS'
  | 'TRACE';

/**
 * Time intervals, e.g. `"5m"` representing 5 minutes.
 */
export type Duration = string;
