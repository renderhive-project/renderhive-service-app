import React, { ReactNode, createContext, useContext, useState } from 'react';

// context type
type LoadingContextType = {
  isLoading: boolean;
  setLoading: React.Dispatch<React.SetStateAction<boolean>>;
};

// create a context
const defaultValue: LoadingContextType = {
  isLoading: true,
  setLoading: () => {}  // Empty function as a placeholder; real function will be provided by the provider.
}
const LoadingContext = createContext<LoadingContextType>(defaultValue);

// create a hook
export const useLoading = () => {
  return useContext(LoadingContext);
}

// create the provider of the context
export const LoadingContextProvider: React.FC<{ children: ReactNode }> = ({ children }) =>  {
  const [isLoading, setLoading] = useState(defaultValue.isLoading);

  return (
    <LoadingContext.Provider value={{ isLoading, setLoading }}>
      {children}
    </LoadingContext.Provider>
  );
};