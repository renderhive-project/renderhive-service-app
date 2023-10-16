import { createContext, ReactNode, useState } from "react";

type MetamaskContextType = {  
  metamaskAccountAddress: string,
  setMetamaskAccountAddress: (newValue: string) => void,
};

const defaultValue: MetamaskContextType = {
  metamaskAccountAddress: '',
  setMetamaskAccountAddress: () => { },
}

export const MetamaskContext = createContext(defaultValue)

export const MetamaskContextProvider = (props: { children: ReactNode | undefined }) => {
  const [metamaskAccountAddress, setMetamaskAccountAddress] = useState('')

  return (
    <MetamaskContext.Provider
      value={{
        metamaskAccountAddress,
        setMetamaskAccountAddress
      }}
    >
      {props.children}
    </MetamaskContext.Provider>
  )
}
