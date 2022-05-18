import React from 'react';
import { GitopsCluster } from '../cluster-services/cluster_services.pb';

export interface FormState {
  activeIndex: number;
  numberOfItems: number;
  cluster: GitopsCluster;
  error: string;
}

export type SetFormState = React.Dispatch<React.SetStateAction<FormState>>;
