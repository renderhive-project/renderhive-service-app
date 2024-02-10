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
      "ContractID": "0.0.3160015",
      "Gas": 30000
    }
  }`,

  // OPERATOR MANAGEMENT
  contractservice_registerOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RegisterOperator",
    "params": {
      "ContractID": "0.0.3160015",
      "Gas": 90000,
      "funcParams": [
        { "type": "string", "name":"_operatorTopic", "value": "0.0.XXXX"}
      ]
    }
  }`,

  contractservice_unregisterOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.UnregisterOperator",
    "params": {
      "ContractID": "0.0.3160015",
      "Gas": 90000,
      "funcParams": []
    }
  }`,

  contractservice_depositOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.DepositOperatorFunds",
    "params": {
      "ContractID": "0.0.3160015",
      "Amount": "1",
      "Gas": 70000,
      "funcParams": []
    }
  }`,

  contractservice_withdrawOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.WithdrawOperatorFunds",
    "params": {
      "ContractID": "0.0.3160015",
      "Gas": 70000,
      "funcParams": [
        { "type": "hbar", "name":"_amount", "value": 1}
      ]
    }
  }`,

  contractservice_getOperatorBalance: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetOperatorBalance",
    "params": {
      "ContractID": "0.0.3160015",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_getReservedOperatorFunds: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetReservedOperatorFunds",
    "params": {
      "ContractID": "0.0.3160015",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_isOperator: `{
    "jsonrpc": "2.0",
    "method": "ContractService.IsOperator",
    "params": {
      "ContractID": "0.0.3160015",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  contractservice_getOperatorLastActivity: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetOperatorLastActivity",
    "params": {
      "ContractID": "0.0.3160015",
      "AccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  // NODE MANAGEMENT
  contractservice_addNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.AddNode",
    "params": {
      "ContractID": "0.0.3160015",
      "Amount": "75",
      "Gas": 230000,
      "funcParams": [
        { "type": "address", "name":"_nodeAccount", "value": "0.0.XXXX"},
        { "type": "string", "name":"_nodeTopic", "value": "0.0.XXXX"}
      ]
    }
  }`,

  contractservice_removeNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.RemoveNode",
    "params": {
      "ContractID": "0.0.3160015",
      "Gas": 180000,
      "funcParams": [
        { "type": "address", "name":"_nodeAccount", "value": "0.0.XXXX"}
      ]
    }
  }`,

  contractservice_isNode: `{
    "jsonrpc": "2.0",
    "method": "ContractService.IsNode",
    "params": {
      "ContractID": "0.0.3160015",
      "NodeAccountID": "0.0.XXXX",
      "OperatorAccountID": "0.0.XXXX",
      "Gas": 40000
    }
  }`,

  contractservice_depositNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.DepositNodeStake",
    "params": {
      "ContractID": "0.0.3160015",
      "Amount": "75",
      "Gas": 100000,
      "funcParams": [
        { "type": "address", "name":"_nodeAccount", "value": "0.0.XXXX"}
      ]
    }
  }`,

  contractservice_withdrawNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.WithdrawNodeStake",
    "params": {
      "ContractID": "0.0.3160015",
      "Gas": 65000,
      "funcParams": [
        { "type": "address", "name":"_nodeAccount", "value": "0.0.XXXX"}
      ]
    }
  }`,

  contractservice_getNodeStake: `{
    "jsonrpc": "2.0",
    "method": "ContractService.GetNodeStake",
    "params": {
      "ContractID": "0.0.3160015",
      "NodeAccountID": "0.0.XXXX",
      "Gas": 30000
    }
  }`,

  // RENDER JOB MANAGEMENT
  contractservice_addRenderJob: `{
    "jsonrpc": "2.0",
    "method": "ContractService.AddRenderJob",
    "params": {
      "ContractID": "0.0.3160015",
      "Amount": "1",
      "Gas": 230000,
      "funcParams": [
        { "type": "string", "name":"_jobCID", "value": ""},
        { "type": "uint256", "name":"_estimatedJobWork", "value": "200"}
      ]
    }
  }`,

  contractservice_claimRenderJob: `{
    "jsonrpc": "2.0",
    "method": "ContractService.ClaimRenderJob",
    "params": {
      "ContractID": "0.0.3160015",
      "JobCID": "",
      "HiveCycle": 1,
      "NodeCount": 1,
      "NodeShare": 1000,
      "ConsensusRoot": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "JobRoot": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "Gas": 230000
    }
  }`,
};

const Playground = () => {
  const [loading, setLoading] = useState(false);
  const { accountId, walletInterface } = useWalletInterface();
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

  // execute the request
  const executeRequest = async () => {

    // define the requests that are processed by the frontend
    const frontend_requests = ['ContractService._isNodeStaked', 'ContractService.RegisterOperator', 'ContractService.UnregisterOperator', 'ContractService.DepositOperatorFunds', 'ContractService.WithdrawOperatorFunds', 'ContractService.AddNode', 'ContractService.RemoveNode', 'ContractService.DepositNodeStake', 'ContractService.WithdrawNodeStake', 'ContractService.AddRenderJob'];

    // get the request body from the editor
    let requestBody = JSON.parse(editorText);

    // if the request method is NOT in the frontend_requests array
    if (!frontend_requests.includes(requestBody.method)) {
        // send the request to the backend
        return sendJsonRpcRequest();

    // execute the request in the frontend
    } else {

        try {

          // set loading and success message
          setLoading(true);
          setSuccessMessage("");

          // get the contract ID
          let contract_id = ContractId.fromString(requestBody.params.ContractID);
          let amount = new Hbar(requestBody.params.Amount);
          let gasLimit = requestBody.params.Gas;

          // get the method name
          let method = requestBody.method;
          let methodName = method.split('.')[1].charAt(0).toLowerCase() + method.split('.')[1].slice(1);

          // prepare the function parameters
          let funcParams = requestBody.params.funcParams;
          let functionParameters = new ContractFunctionParameterBuilder();
          
          // if there are function parameters
          if (funcParams) {
        
              // loop through all parameters, convert them to the correct format, and add them to the functionParameters object
              let param, type, value;
              for (let i = 0; i < funcParams.length; i++) {
                  param = funcParams[i];
                  type = param.type;
                  value = param.value;
                  
                  // convert parameters to the correct format
                  if (param.type === "address") {
                      let account_id = AccountId.fromString(param.value);
                      value = account_id.toSolidityAddress();
                  } else if (param.type === "hbar") {
                      let hbar = new Hbar(param.value);
                      type = "uint256";
                      value = hbar.toTinybars();
                  }

                  // add parameter to the transaction's function parameters
                  functionParameters = functionParameters.addParam({type:type, name:param.name, value:value});
              }
          }

          // prepare the hedera transaction to call the contract
          let response_contractexecute_txnId;
          if (amount !== undefined) {
              response_contractexecute_txnId = await walletInterface.executeContractFunction(contract_id, methodName, functionParameters, gasLimit, amount);
          } else {
              response_contractexecute_txnId = await walletInterface.executeContractFunction(contract_id, methodName, functionParameters, gasLimit);
          }
          console.log(response_contractexecute_txnId);
          if (response_contractexecute_txnId) {

              // get the contract call's results form a mirror node using the transaction id



              console.log(response_contractexecute_txnId);
              setSuccessMessage(methodName + " function was called with transaction: " + response_contractexecute_txnId.toString());
          }
          // // request signing of the data
          // const response_createaccount_txnId = await walletInterface.transferHBAR(AccountId.fromString(response_init.data.result.NodeAccountID), 10)
          // console.log(response_createaccount_txnId)
          // if (response_createaccount_txnId) {

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
        </Grid>

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
      </Box>

    </Box>

  )
}

export default Playground;