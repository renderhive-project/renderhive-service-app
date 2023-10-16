import { createContext, useState, ReactNode } from "react";

type HashconnectContextType = {  
  accountId: string,
  setAccountId: (newValue: string) => void,
  isConnected: boolean,
  setIsConnected: (newValue: boolean) => void,
};

const defaultValue: HashconnectContextType = {
  accountId: '',
  setAccountId: () => { },
  isConnected: false,
  setIsConnected: () => { },
}

export const HashconnectContext = createContext(defaultValue);

export const HashconnectContextProvider = (props: { children: ReactNode | undefined }) => {
  const [accountId, setAccountId] = useState<string>(defaultValue.accountId);
  const [isConnected, setIsConnected] = useState<boolean>(defaultValue.isConnected);

  return (
    <HashconnectContext.Provider
      value={{
        accountId,
        setAccountId,
        isConnected,
        setIsConnected
      }}
    >
      {props.children}
    </HashconnectContext.Provider>
  )
}
