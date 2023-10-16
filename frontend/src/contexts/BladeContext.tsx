import { createContext, ReactNode, useState } from "react";

type BladeContextType = {  
  accountId: string,
  setAccountId: (newValue: string) => void,
  isConnected: boolean,
  setIsConnected: (newValue: boolean) => void,
};

const defaultValue: BladeContextType = {
  accountId: '',
  setAccountId: () => { },
  isConnected: false,
  setIsConnected: () => { },
}

export const BladeContext = createContext(defaultValue)

export const BladeContextProvider = (props: { children: ReactNode | undefined }) => {
  const [accountId, setAccountId] = useState<string>(defaultValue.accountId);
  const [isConnected, setIsConnected] = useState<boolean>(defaultValue.isConnected);
  return (
    <BladeContext.Provider
      value={{
        accountId,
        setAccountId,
        isConnected,
        setIsConnected
      }}
    >
      {props.children}
    </BladeContext.Provider>
  )
}
