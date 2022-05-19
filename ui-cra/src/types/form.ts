import React from 'react';
import { GitopsCluster } from '../capi-server/capi_server.pb';

export interface FormState {
  activeIndex: number;
  numberOfItems: number;
  cluster: GitopsCluster;
  error: string;
}

export type SetFormState = React.Dispatch<React.SetStateAction<FormState>>;
