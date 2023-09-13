import React, { useState } from 'react';
import axios from 'axios';
import { appConfig } from "../../config";

const RPCTest = () => {
    const [identifier, setIdentifier] = useState('');
    const [operatorInfo, setOperatorInfo] = useState(null);
    const [errorMessage, setErrorMessage] = useState('');

    const handleGetInfo = async () => {
        try {
            const response = await axios.post(appConfig.api.jsonrpc.url, {
                jsonrpc: '2.0',
                method: 'OperatorService.SignUp',
                params: [{ 
                    Operator: { 
                        Username: "Test", 
                        Email: "Test@domain.io", 
                        AccountID: "0.0.390079" 
                    } 
                }],
                id: 1
            });

            if (response.data && response.data.result) {
                // Handle successful response
                setOperatorInfo(response.data.result);
                setErrorMessage(''); // clear any previous error messages
                console.log(response.data.result);
            } else {
                setOperatorInfo(null);
                setErrorMessage('Unexpected server response.');
                console.log("Unexpected response format:", response.data);
            }

        } catch (error: any) {
            if (error.response && error.response.data && error.response.data.error) {
                // Handle JSON-RPC error sent with an HTTP 400 status code
                setOperatorInfo(null);
                setErrorMessage(error.response.data.error);
            } else {
                // Handle other types of errors
                setOperatorInfo(null);
                setErrorMessage('Failed to connect to the server.');
            }
            console.error("Error calling the API", error);
        }
    };

    return (
        <div className="App">
            <input 
                value={identifier}
                onChange={e => setIdentifier(e.target.value)}
                placeholder="Enter operator identifier"
            />
            <button onClick={handleGetInfo}>Get Operator Info</button>
            {errorMessage ? (
                <div className="error">{errorMessage}</div>
            ) : (
                operatorInfo && (
                    <div>
                        <strong>Username:</strong> {operatorInfo.Username}
                        <br />
                        <strong>Email:</strong> {operatorInfo.Email}
                        <br />
                        <strong>Account:</strong> {operatorInfo.Account}
                    </div>
                )
            )}
        </div>
    );
}

export default RPCTest;