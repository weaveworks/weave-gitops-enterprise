import { getCallbackState } from '@weaveworks/weave-gitops';
import { useState } from 'react';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  return callbackState;
};
