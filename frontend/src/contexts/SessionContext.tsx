import { createContext, useContext, useEffect, useState } from "react";
import { appConfig } from "../config";
import { useWalletInterface } from "../services/wallets/useWalletInterface";

import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// define types we need
type OperatorInfo = {
    accountId: string;
    username: string;
    email: string;
}
type NodeInfo = {
    accountId: string;
    name: string;
};

type SessionContextType = {
    signedIn: boolean,
    setSignedIn: (newValue: boolean) => void,
    operatorInfo: OperatorInfo,
    setOperatorInfo: (newValue: OperatorInfo) => void; // Define the type for newValue
    nodeInfo: NodeInfo,
    setNodeInfo: (newValue: NodeInfo) => void;  // Define the type for newValue
};

const defaultValue: SessionContextType = {
    signedIn: false,
    setSignedIn: (newValue: boolean) => { },
    operatorInfo: {
        accountId: '',
        username: '',
        email: ''
    },
    setOperatorInfo: (newValue: OperatorInfo) => { },
    nodeInfo: {
        accountId: '',
        name: '',
    },
    setNodeInfo: (newValue: NodeInfo) => { },
}

export const SessionContext = createContext<SessionContextType>(defaultValue);

export const useSession = () => {
  const context = useContext(SessionContext);
  if (!context) throw new Error("useSession must be used within a SessionProvider");
  return context;
};

export const SessionContextProvider = ({ children }) => {
  const {accountId, walletInterface} = useWalletInterface();
  const [signedIn, setSignedIn] = useState(defaultValue.signedIn);
  const [operatorInfo, setOperatorInfo] = useState(defaultValue.operatorInfo);
  const [nodeInfo, setNodeInfo] = useState(defaultValue.nodeInfo);

  // get info about the operator and the node
  useEffect(() => {

    const fetchInfo = async () => {
      try {

        // Request basic node infos
        const response = await axios.post(appConfig.api.jsonrpc.url, {
          jsonrpc: '2.0',
          method: 'OperatorService.GetInfo',
          params: [{}],
          id: uuidv4()
        }, {
          headers: {
              'Content-Type': 'application/json',
          },
          withCredentials: true, // This will include cookies in the request
        })

        // set operator and node data
        if (response.data && response.data.result) {
            setOperatorInfo({
              username: response.data.result.Username,
              email: response.data.result.UserEmail,
              accountId: response.data.result.UserAccount,
            })
            setNodeInfo({
              name: response.data.result.NodeName,
              accountId: response.data.result.NodeAccount,
            })

        } else {
            //setErrorMessage('Unexpected server response.');
            console.error("Unexpected response format:", response.data);
        }

          
      } catch (error) {
          console.error('Error signing in:', error);
          //setErrorMessage('Error signing in. Please try again.');
      }
    };

    // if the accountId is not empty anymore
    if (accountId) {
      fetchInfo();
    }

  }, [accountId])
  
  // get info about the operator and the node
  useEffect(() => {

    const checkSession = async () => {
        try {

          // Check the session validity
          const response = await axios.post(appConfig.api.jsonrpc.url, {
            jsonrpc: '2.0',
            method: 'OperatorService.IsSessionValid',
            params: [{}],
            id: uuidv4()
          }, {
            headers: {
                'Content-Type': 'application/json',
            },
            withCredentials: true, // This will include cookies in the request
          })

          // if a response was received
          if (response.data && response.data.result) {
              setSignedIn(response.data.result.Valid)

          } else {
              //setErrorMessage('Unexpected server response.');
              console.error("Unexpected response format:", response.data);
              setSignedIn(false)
          }

            
        } catch (error) {
            console.error('Error signing in:', error);
            //setErrorMessage('Error signing in. Please try again.');
        }
      };

      // if the accountId is not empty anymore
      checkSession();

  }, [])

  const contextValue = {
      signedIn,
      setSignedIn,
      operatorInfo,
      setOperatorInfo,
      nodeInfo,
      setNodeInfo,
    };

  return (
    <SessionContext.Provider value={contextValue}>
      {children}
    </SessionContext.Provider>
  );
};