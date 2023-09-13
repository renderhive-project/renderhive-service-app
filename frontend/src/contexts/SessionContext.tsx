import { createContext, useEffect, useState } from "react";
import axios from 'axios';
import { appConfig } from "../config";

// define types we need
type OperatorInfo = {
    accountId: string;
    username: string;
    email: string;
}
type NodeInfo = {
    accountId: string;
    nodename: string;
    balance: string;
};

type SessionContextType = {
    operatorInfo: OperatorInfo,
    setOperatorInfo: (newValue: OperatorInfo) => void; // Define the type for newValue
    nodeInfo: NodeInfo,
    setNodeInfo: (newValue: NodeInfo) => void;  // Define the type for newValue
};

const defaultValue: SessionContextType = {
    operatorInfo: {
        accountId: '',
        username: '',
        email: ''
    },
    setOperatorInfo: (newValue: OperatorInfo) => { },
    nodeInfo: {
        accountId: '',
        nodename: '',
        balance: ''
    },
    setNodeInfo: (newValue: NodeInfo) => { },
}

const SessionContext = createContext<SessionContextType>(defaultValue);

export const SessionContextProvider = ({ children }) => {
    const [operatorInfo, setOperatorInfo] = useState(defaultValue.operatorInfo);
    const [nodeInfo, setNodeInfo] = useState(defaultValue.nodeInfo);

    const contextValue = {
        operatorInfo,
        setOperatorInfo,
        nodeInfo,
        setNodeInfo,
      };
  
    // // fetch session data from the backend
    // useEffect(() => {
    //   async function fetchData() {
    //     try {
    //         const response = await axios.post(appConfig.api.jsonrpc.url, {
    //             jsonrpc: '2.0',
    //             method: 'OperatorService.SignUp',
    //             params: [{ 
    //                 Operator: { 
    //                     Username: "Test", 
    //                     Email: "Test@domain.io", 
    //                     AccountID: "0.0.390079" 
    //                 } 
    //             }],
    //             id: 1
    //         });

    //         if (response.data && response.data.result) {
    //             // Handle successful response
    //             setOperatorInfo(response.data.result);
    //             setErrorMessage(''); // clear any previous error messages
    //             console.log(response.data.result);
    //         } else {
    //             setOperatorInfo(null);
    //             setErrorMessage('Unexpected server response.');
    //             console.log("Unexpected response format:", response.data);
    //         }

    //         // store the session data
    //         setSessionData(sessionData)

    //     } catch (error: any) {
    //         if (error.response && error.response.data && error.response.data.error) {
    //             // Handle JSON-RPC error sent with an HTTP 400 status code
    //             setOperatorInfo(null);
    //             setErrorMessage(error.response.data.error);
    //         } else {
    //             // Handle other types of errors
    //             setOperatorInfo(null);
    //             setErrorMessage('Failed to connect to the server.');
    //         }
    //         console.error("Error calling the API", error);
    //     }
    //   }
  
    //   fetchData();
    // }, []);
  
    return (
      <SessionContext.Provider value={contextValue}>
        {children}
      </SessionContext.Provider>
    );
  };