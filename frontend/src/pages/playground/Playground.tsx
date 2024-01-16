import { appConfig } from "../../config";
import { useState, useEffect } from 'react';
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

  // OPERATOR MANAGEMENT
  contractservice_registerOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RegisterOperator",
    "params": {
      "ContractID": "0.0.XXXX",
      "OperatorTopicID": "0.0.XXXX",
      "Gas": 90000
    }
  }`,

  contractservice_unregisterOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.UnregisterOperator",
    "params": {
      "ContractID": "0.0.XXXX",
      "Gas": 90000
    }
  }`,

  contractservice_depositOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.DepositOperatorFunds",
    "params": {
      "ContractID": "0.0.XXXX",
      "Amount": "1",
      "Gas": 70000
    }
  }`,

  contractservice_withdrawOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.WithdrawOperatorFunds",
    "params": {
      "ContractID": "0.0.XXXX",
      "Amount": "1",
      "Gas": 70000
    }
  }`,

  contractservice_getOperatorBalance: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetOperatorBalance",
    "params": {
      "ContractID": "0.0.XXXX",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_getReservedOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetReservedOperatorFunds",
    "params": {
      "ContractID": "0.0.XXXX",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_isOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.IsOperator",
    "params": {
      "ContractID": "0.0.XXXX",
      "AccountID": "0.0.XXXX",
      "Gas": 22000
    }
  }`,

  // NODE MANAGEMENT
  contractservice_addNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.AddNode",
    "params": {
      "ContractID": "0.0.XXXX",
      "AccountID": "0.0.XXXX",
      "TopicID": "0.0.XXXX",
      "Gas": 230000
    }
  }`,

  contractservice_removeNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RemoveNode",
    "params": {
      "ContractID": "0.0.XXXX",
      "AccountID": "0.0.XXXX",
      "Gas": 230000
    }
  }`,
};

const Playground = () => {
  const [loading, setLoading] = useState(false);
  const { accountId } = useWalletInterface();
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [successMessage, setSuccessMessage] = useState<string>('');
  const { signedIn, setSignedIn, operatorInfo, nodeInfo } = useSession();
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
  const sendJsonRpcRequest = async () => {
    try {
      setLoading(true);
      setSuccessMessage("");
  
      // get the request body from the editor
      let requestBody = JSON.parse(editorText);

      // Add the id and timeout fields to the request body
      requestBody = {
        ...requestBody,
        id: uuidv4(),
        timeout: appConfig.constants.BACKEND_JSONRPC_TIMEOUT,
      };
  
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


  return (
    <Box className="playground">

      {/* Box 1 */}
      <Box width={{ xs: '100%', sm: '90%', md: '80%', lg: '80%' }}>

        {/* Page Title */}
        <Typography variant="h4" sx={{ marginBottom: '10px', }}>JSON-RPC Playground</Typography>

        {/* Content */}
        <Grid container spacing={0} className="page-content" bgcolor="background.paper">
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
                    <MenuItem value="contractservice_registerOperator">Register Operator</MenuItem>
                    <MenuItem value="contractservice_unregisterOperator">Delete Operator</MenuItem>
                    <MenuItem value="contractservice_depositOperatorFunds">Deposit HBAR</MenuItem>
                    <MenuItem value="contractservice_withdrawOperatorFunds">Withdraw HBAR</MenuItem>
                    <MenuItem value="contractservice_getOperatorBalance">Get Operator Balance</MenuItem>
                    <MenuItem value="contractservice_getReservedOperatorFunds">Get Reserved Funds</MenuItem>
                    <MenuItem value="contractservice_isOperator">Verify Operator</MenuItem>
                    <Divider />
                    <MenuItem value="contractservice_addNode">Add Node</MenuItem>
                    <MenuItem value="contractservice_removeNode">Remove Node</MenuItem>
                    {/* Add more menu items as needed */}
                  </Select>
                </Box>

                {/* Send Request */}
                <LoadingButton
                  fullWidth
                  onClick={sendJsonRpcRequest}
                  setError={setErrorMessage}
                  loadingText="Sending ..."
                  buttonText="Send Request"
                />

                {/* Error messages */}
                {errorMessage && (
                  <Box marginTop={2}>
                    <Alert severity="error" onClose={() => setErrorMessage("")} sx={{ marginBottom: '10px', marginTop: '10px' }}>
                        {errorMessage}
                    </Alert>
                  </Box>
                )}

                {/* Success messages */}
                {successMessage && (
                  <Box marginTop={2}>
                    <Alert severity="success" onClose={() => setSuccessMessage("")} sx={{ marginBottom: '10px', marginTop: '10px' }}>
                        {successMessage}
                    </Alert>
                  </Box>
                )}
              </Grid>
            </Grid>
          </Grid>
        </Grid>

      </Box>

    </Box>

  )
}

export default Playground;