import { clearCallbackState, getCallbackState } from '@weaveworks/weave-gitops';
import { useState } from 'react';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  clearCallbackState();
  return callbackState;
};
