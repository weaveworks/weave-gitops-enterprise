import React from 'react';
import { Cluster } from './kubernetes';

export interface FormState {
  activeIndex: number;
  numberOfItems: number;
  cluster: Cluster;
  error: string;
}

export type SetFormState = React.Dispatch<React.SetStateAction<FormState>>;
