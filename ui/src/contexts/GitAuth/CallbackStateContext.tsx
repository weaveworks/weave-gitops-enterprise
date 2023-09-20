import { CallbackSessionState } from '../../components/GitAuth/utils';
import * as React from 'react';

type Props = {
  callbackState: CallbackSessionState;
  children?: any;
};

export type CallbackStateContextType = {
  callbackState: CallbackSessionState;
};
export const CallbackStateContext =
  React.createContext<CallbackStateContextType | null>(null);

export default function CallbackStateContextProvider({
  callbackState,
  children,
}: Props) {
  const value: CallbackStateContextType = {
    callbackState,
  };
  return <CallbackStateContext.Provider value={value} children={children} />;
}
