import { useState } from 'react';
import { getCallbackState } from '../components/GitAuth/utils';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  return callbackState;
};
