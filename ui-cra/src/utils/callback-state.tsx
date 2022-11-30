import { useState } from 'react';
import { getCallbackState } from '../contexts/GithubAuth/utils';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  return callbackState;
};
