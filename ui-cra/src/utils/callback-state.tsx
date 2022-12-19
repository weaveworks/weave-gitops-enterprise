import { useState } from 'react';
import { getCallbackState } from '../components/GithubAuth/utils';

export const useCallbackState = () => {
  const [callbackState] = useState(getCallbackState());
  return callbackState;
};
