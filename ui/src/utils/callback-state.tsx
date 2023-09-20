import { getCallbackState } from '../components/GitAuth/utils';
import { useState } from 'react';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  return callbackState;
};
