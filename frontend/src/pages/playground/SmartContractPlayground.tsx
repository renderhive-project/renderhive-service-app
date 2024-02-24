import { appConfig } from "../../config";
import { useState, useEffect } from 'react';
import Linkify from 'react-linkify';
import { useSession } from '../../contexts/SessionContext';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// components
import { Alert, Box, Divider, Grid, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material'
import LoadingButton from "../../components/button/LoadingButton";

// Code editor
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-github_dark";

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { AccountId, ContractId, Hbar } from "@hashgraph/sdk";
import { ContractFunctionParameterBuilder } from "../..//services//wallets/contractFunctionParameterBuilder";

// styles
import "./playground.scss"

// define JSON-RPC call templates
const jsonRpcTemplates = {
  // GENERAL
  contractservice_deploy: `{
    "jsonrpc": "2.0",
    "method": "ContractService.Deploy",
    "params": {
      "ContractFilepath": "./RenderhiveContract.bin",
      "Gas": 15000000
    }
  }`,

  contractservice_getCurrentHiveCyle: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetCurrentHiveCycle",
    "params": {
      "ContractID": "0.0.3566638",
      "Gas": 30000
    }
  }`,

  // OPERATOR MANAGEMENT
  contractservice_registerOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RegisterOperator",
    "params": {
        "ContractID": "0.0.3566638",
        "OperatorTopic": "0.0.XXXX",
        "Gas": 90000
    }
  }`,

  contractservice_unregisterOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.UnregisterOperator",
    "params": {
        "ContractID": "0.0.3566638",
        "Gas": 90000
    }
  }`,

  contractservice_depositOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.DepositOperatorFunds",
    "params": {
        "ContractID": "0.0.3566638",
        "Amount": "1",
        "Gas": 70000
    }
  }`,

  contractservice_withdrawOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.WithdrawOperatorFunds",
    "params": {
      "ContractID": "0.0.3566638",
      "Amount": "1",
      "Gas": 70000
    }
  }`,

  contractservice_getOperatorBalance: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetOperatorBalance",
    "params": {
      "ContractID": "0.0.3566638",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_getReservedOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetReservedOperatorFunds",
    "params": {
      "ContractID": "0.0.3566638",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_isOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.IsOperator",
    "params": {
      "ContractID": "0.0.3566638",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_getOperatorLastActivity: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetOperatorLastActivity",
    "params": {
      "ContractID": "0.0.3566638",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  // NODE MANAGEMENT
  contractservice_addNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.AddNode",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "TopicID": "0.0.XXXX",
      "NodeStake": "75",
      "Gas": 230000
    }
  }`,

  contractservice_removeNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RemoveNode",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "Gas": 180000
    }
  }`,

  contractservice_isNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.IsNode",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "OperatorAccountID": "0.0.XXXX",
      "Gas": 40000
    }
  }`,

  contractservice_depositNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.DepositNodeStake",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "NodeStake": "75",
      "Gas": 100000
    }
  }`,

  contractservice_withdrawNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.WithdrawNodeStake",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "Gas": 65000
    }
  }`,

  contractservice_getNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetNodeStake",
    "params": {
      "ContractID": "0.0.3566638",
      "NodeAccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  // RENDER JOB MANAGEMENT
  contractservice_addRenderJob: `{
    "jsonrpc": "2.0",
    "method": "ContractService.AddRenderJob",
    "params": {
      "ContractID": "0.0.3566638",
      "JobCID": "",
      "Work": 200,
      "Funding": "1",
      "Gas": 230000
    }
  }`,

  contractservice_claimRenderJob: `{
    "jsonrpc": "2.0",
    "method": "ContractService.ClaimRenderJob",
    "params": {
      "ContractID": "0.0.3566638",
      "JobCID": "",
      "HiveCycle": 1,
      "NodeCount": 1,
      "NodeShare": 1000,
      "ConsensusRoot": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "JobRoot": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "Gas": 150000
    }
  }`,
}

const SmartContractPlayground = () => {
    const [loading, setLoading] = useState(false);
    const { accountId, walletInterface } = useWalletInterface();
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [successMessage, setSuccessMessage] = useState<string>('');
    const [editorText, setEditorText] = useState<string>('');
  
    // set the default template
    const defaultTemplateKey = "contractservice_deploy";
    useEffect(() => {
      const selectedTemplate = jsonRpcTemplates[defaultTemplateKey];
      const formattedTemplate = JSON.stringify(JSON.parse(selectedTemplate), null, 2);
      setEditorText(formattedTemplate);
    }, []);
  
    const handleTemplateChange = (event: SelectChangeEvent<string>) => {
      const selectedTemplate = jsonRpcTemplates[event.target.value as keyof typeof jsonRpcTemplates];
      const formattedTemplate = JSON.stringify(JSON.parse(selectedTemplate), null, 2);
      setEditorText(formattedTemplate);
    };
  
    // send a request to the backend (JSON-RPC)
    const sendJsonRpcRequest = async (requestBody: object) => {
      try {
        setLoading(true);
        setSuccessMessage("");
    
        // add headers, credentials, and send the request
        const response = await axios.post(appConfig.api.jsonrpc.url, requestBody, {
          headers: {
            'Content-Type': 'application/json',
          },
          withCredentials: true,
        });
    
        if (response.data && response.data.result) {
          console.log(response.data.result);
          setSuccessMessage(response.data.result.Message);
        } else {
          console.log()
          setLoading(false);
  
          throw new Error(response.data.error.message);
        }
      } catch (error) {
        setLoading(false);
        console.error('Error:', error);
        return Promise.reject(error);
      } finally {
        setLoading(false);
      }
    };
  
    // execute the request
    const executeRequest = async () => {

        // define the requests that are processed by the frontend
        const frontend_requests = ['ContractService._isNodeStaked', 'ContractService.RegisterOperator', 'ContractService.UnregisterOperator', 'ContractService.DepositOperatorFunds', 'ContractService.WithdrawOperatorFunds', 'ContractService.AddNode', 'ContractService.RemoveNode', 'ContractService.DepositNodeStake', 'ContractService.WithdrawNodeStake', 'ContractService.AddRenderJob'];

        // get the request body from the editor
        let requestBody = JSON.parse(editorText);

        // Add the id and timeout fields to the request body
        requestBody = {
            ...requestBody,
            id: uuidv4(),
            timeout: appConfig.constants.BACKEND_JSONRPC_TIMEOUT,
        };

        // if the request method is NOT in the frontend_requests array
        if (!frontend_requests.includes(requestBody.method)) {
            // send the request to the backend
            return sendJsonRpcRequest(requestBody);

        // execute the request in the frontend
        } else {

            try {

            // set loading and success message
            setLoading(true);
            setSuccessMessage("");

            // add headers, credentials, and send the request
            const response = await axios.post(appConfig.api.jsonrpc.url, requestBody, {
                headers: {
                'Content-Type': 'application/json',
                },
                withCredentials: true,
            });
            console.log(response.data.result);
        
            if (response.data && response.data.result) {
                setSuccessMessage(response.data.result.Message);

                // if a transaction in bytes is returned, send it to the Hedera network
                if (response.data.result.TransactionBytes == undefined) {
                    setLoading(false);
                    throw new Error("Failed to retrieve transaction bytes from the backend");
                }

                // send the transaction to the Hedera network
                const transactionBytes = response.data.result.TransactionBytes;
                const response_execute_txnId = await walletInterface.executeTransaction(transactionBytes);
                // if the transaction was executed successfully
                if (response_execute_txnId) {

                    // get the contract call's results form a mirror node using the transaction id
                    console.log(response_execute_txnId);
                    setSuccessMessage("Successfully executed transaction: http://hashscan.io/testnet/transaction/" + response_execute_txnId);
                }

            } else {
                console.log()
                setLoading(false);
        
                throw new Error(response.data.error.message);
            }

            } catch (error) {
            setLoading(false);
            console.error('Error:', error);
            return Promise.reject(error);

            } finally {
            setLoading(false);

            }
        }
  
    };

    return (
        <>
            {/* Smart Contract Playground */}
            <Grid container spacing={0} className="page-content" bgcolor="background.paper">

                {/* Title */}
                <Typography variant="h5" sx={{ marginBottom: '10px', }}>Smart Contract</Typography>
                
                <Grid item xs={12} sx={{height: '500px'}}>
                <Grid container spacing={4}>
                    <Grid item xs={9}>
                    <AceEditor
                        mode="json"
                        theme="github_dark"
                        onChange={setEditorText}
                        name="rpc-request-editor"
                        editorProps={{ $blockScrolling: true }}
                        value={editorText}
                        style={{ width: '100%' }}
                    />
                    </Grid>
                    <Grid item xs={3}>
                    <Box marginBottom={2}>

                        {/* Select Request Template */}
                        <Select
                        defaultValue={defaultTemplateKey}
                        onChange={handleTemplateChange}
                        fullWidth
                        >
                        <MenuItem value="contractservice_deploy">Deploy Smart Contract</MenuItem>
                        <MenuItem value="contractservice_getCurrentHiveCyle">Get Current Hive Cycle</MenuItem>
                        <Divider />
                        <MenuItem value="contractservice_registerOperator">Register Operator</MenuItem>
                        <MenuItem value="contractservice_unregisterOperator">Delete Operator</MenuItem>
                        <MenuItem value="contractservice_depositOperatorFunds">Deposit HBAR</MenuItem>
                        <MenuItem value="contractservice_withdrawOperatorFunds">Withdraw HBAR</MenuItem>
                        <MenuItem value="contractservice_getOperatorBalance">Get Operator Balance</MenuItem>
                        <MenuItem value="contractservice_getReservedOperatorFunds">Get Reserved Funds</MenuItem>
                        <MenuItem value="contractservice_isOperator">Verify Operator</MenuItem>
                        <MenuItem value="contractservice_getOperatorLastActivity">Get Last Operator Activity</MenuItem>
                        <Divider />
                        <MenuItem value="contractservice_addNode">Add Node</MenuItem>
                        <MenuItem value="contractservice_removeNode">Remove Node</MenuItem>
                        <MenuItem value="contractservice_isNode">Verify Node</MenuItem>
                        <MenuItem value="contractservice_depositNodeStake">Deposit Node Stake</MenuItem>
                        <MenuItem value="contractservice_withdrawNodeStake">Withdraw Node Stake</MenuItem>
                        <MenuItem value="contractservice_getNodeStake">Get Node Stake</MenuItem>
                        <Divider />
                        <MenuItem value="contractservice_addRenderJob">Add Render Job</MenuItem>
                        <MenuItem value="contractservice_claimRenderJob">Claim Render Job</MenuItem>
                        {/* Add more menu items as needed */}
                        </Select>
                    </Box>

                    {/* Send Request */}
                    <LoadingButton
                        fullWidth
                        onClick={executeRequest}
                        setError={setErrorMessage}
                        loadingText="Sending ..."
                        buttonText="Send Request"
                    />

                    </Grid>
                </Grid>
                </Grid>

                {/* Error messages */}
                {errorMessage && (
                <Box marginTop={2} width="100%">
                    <Alert severity="error" onClose={() => setErrorMessage("")} sx={{ marginBottom: '10px', marginTop: '10px' }}>
                    <Linkify componentDecorator={(href, text, key) => (
                        <a href={href} key={key} style={{ textDecoration: 'underline' }}>
                          {text}
                        </a>
                      )}>
                        {errorMessage}
                      </Linkify>
                    </Alert>
                </Box>
                )}

                {/* Success messages */}
                {successMessage && (
                <Box marginTop={2} width="100%">
                    <Alert severity="success" onClose={() => setSuccessMessage("")} sx={{ marginBottom: '10px', marginTop: '10px' }}>
                      <Linkify componentDecorator={(href, text, key) => (
                        <a href={href} key={key} style={{ textDecoration: 'underline' }}>
                          {text}
                        </a>
                      )}>
                        {successMessage}
                      </Linkify>
                    </Alert>
                </Box>
                )}

            </Grid>
        </>
    )

}

export default SmartContractPlayground;