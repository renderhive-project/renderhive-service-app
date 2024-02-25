import { appConfig } from "../../config";
import { useState, useEffect, useCallback } from 'react';
import Linkify from 'react-linkify';
import { useSession } from '../../contexts/SessionContext';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// components
import { Alert, Box, Button, Divider, Grid, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material'
import { useDropzone } from 'react-dropzone';
import LoadingButton from "../../components/button/LoadingButton";
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

// Code editor
import AceEditor from 'react-ace';
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-github_dark";

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';

// Icons
import CloudUploadIcon from '@mui/icons-material/CloudUpload';

// styles
import "./playground.scss"
import { error } from "console";

// define JSON-RPC call templates
const jsonRpcTemplates = {
  // GENERAL
  renderhiveservice_createrenderrequest: `{
    "jsonrpc": "2.0",
    "method": "NodeService.CreateRenderRequest",
    "params": {
      "blender": {
        "version": "4.00",
        "engine": "CYCLES",
        "device": "GPU"
      },
      "price": 0.1
    }
  }`,
  renderhiveservice_submitrenderrequest: `{
    "jsonrpc": "2.0",
    "method": "NodeService.SubmitRenderRequest",
    "params": {
      "RenderRequestCID": ""
    }
  }`,
  renderhiveservice_cancelrenderrequest: `{
    "jsonrpc": "2.0",
    "method": "NodeService.CancelRenderRequest",
    "params": {
      "RenderRequestCID": ""
    }
  }`,

  // "resolutionX": "1920",
  // "resolutionY": "1080",
  // "samples": "128",
  // "frameStart": 1,
  // "frameEnd": 250,
  // "frameStep": 1,
  // "outputFormat": "PNG",
  // "outputPath": "/render/output",
  // "outputFilename": "render"
}

const RenderRequestPlayground = () => {
    const [loading, setLoading] = useState(false);
    const { accountId, walletInterface } = useWalletInterface();
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [successMessage, setSuccessMessage] = useState<string>('');
    const [editorText, setEditorText] = useState<string>('');
  
    // set the default template
    const defaultTemplateKey = "renderhiveservice_createrenderrequest";
    useEffect(() => {
      const selectedTemplate = jsonRpcTemplates[defaultTemplateKey];
      const formattedTemplate = JSON.stringify(JSON.parse(selectedTemplate), null, 2);
      setEditorText(formattedTemplate);
    }, []);
  
    const handleTemplateChange = (event: SelectChangeEvent<string>) => {
      const selectedTemplate = jsonRpcTemplates[event.target.value as keyof typeof jsonRpcTemplates];
      const formattedTemplate = JSON.stringify(JSON.parse(selectedTemplate), null, 2);
      setEditorText(formattedTemplate);
      console.log(formattedTemplate);
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
      const frontend_requests = ['NodeService.SubmitRenderRequest', 'NodeService.CancelRenderRequest'];
  
      // get the request body from the editor
      let requestBody = JSON.parse(editorText);
  
      // if this is the create render request command
      if (requestBody.method === "NodeService.CreateRenderRequest") {

          // serialize the files to base64 and add them to the request body
          const files = await Promise.all(
            requestFiles.map((file) => 
              new Promise<{fileName: string, fileScheme: string, fileData: string}>((resolve, reject) => {
                const reader = new FileReader();
                reader.onload = () => {
                  const result = reader.result as string;
                  const urlScheme = result.split(',')[0]; // get the urlScheme
                  const base64Data = result.split(',')[1]; // get the base64 data
                  resolve({ fileName: file.name, fileScheme: urlScheme, fileData: base64Data });
                };
                reader.onerror = reject;
                reader.readAsDataURL(file);
              })
            )
          );

          // Add the files array to the params field of the request body
          requestBody.params = {
            ...requestBody.params,
            files: files,
          };

      }
      
      // Add the id and timeout fields to the request body
      requestBody = {
        ...requestBody,
        id: uuidv4(),
        timeout: appConfig.constants.BACKEND_JSONRPC_TIMEOUT,
      };

      console.log(requestBody);

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

    // handle file drop in the dropzone
    const [requestFiles, setRequestFiles] = useState<File[]>([]);
    const { getRootProps, getInputProps } = useDropzone({
      onDrop: (acceptedFiles) => {
        setRequestFiles(acceptedFiles);
      },
    });

    const fileTable = requestFiles.length > 0 && (
      <Box width="100%" bgcolor="#050F15" sx={{ padding: '10px'}}>
        <TableContainer component={Box} width="100%" bgcolor="#050F15" sx={{ padding: '0px'}}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell style={{ fontSize: '0.8rem', borderBottom: '1px solid grey', padding: '0px' }}>Name</TableCell>
                <TableCell style={{ fontSize: '0.8rem', borderBottom: '1px solid grey', padding: '0px' }} align="right">Size (kB)</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {requestFiles.map((file) => (
                <TableRow key={file.name}>
                  <TableCell component="th" scope="row" style={{ fontSize: '0.8rem', borderBottom: 'none', padding: '0px' }}>
                    {file.name}
                  </TableCell>
                  <TableCell align="right" style={{ fontSize: '0.8rem', borderBottom: 'none', padding: '0px' }}>{file.size / 1000}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    )

    return (
        <>
            {/* Render Request Playground */}
            <Grid container spacing={0} className="page-content" bgcolor="background.paper">

                {/* Title */}
                <Typography variant="h5" sx={{ marginBottom: '10px', }}>Render Requests</Typography>
                
                <Grid item xs={12}>
                  <Grid container spacing={4}>
                    <Grid item xs={9}>
                      <AceEditor
                          mode="json"
                          theme="github_dark"
                          onChange={setEditorText}
                          name="rpc-request-editor"
                          editorProps={{ $blockScrolling: true }}
                          value={editorText}
                          style={{ width: '100%', marginBottom: '15px'}}
                      />

                      {/* File Dropzone */}
                      <Box {...getRootProps({className: 'dropzone'})} style={{ marginBottom: '15px', padding: '0px'}}>
                        {fileTable}
                        <Box width="100%" sx={{ padding: '10px'}}>
                          <input {...getInputProps()} />
                          <span style={{ display: 'flex', alignItems: 'center', fontSize: 'smaller' }}>
                            <CloudUploadIcon sx={{marginRight: '15px'}}/> Add your Blender project files here
                          </span>
                        </Box>
                      </Box>

                    </Grid>
                    <Grid item xs={3}>
                      <Box marginBottom={2}>

                          {/* Select Request Template */}
                          <Select
                          defaultValue={defaultTemplateKey}
                          onChange={handleTemplateChange}
                          fullWidth
                          >
                          <MenuItem value="renderhiveservice_createrenderrequest">Create Render Request</MenuItem>
                          <MenuItem value="renderhiveservice_submitrenderrequest">Submit Render Request</MenuItem>
                          <MenuItem value="renderhiveservice_cancelrenderrequest">Cancel Render Request</MenuItem>
                          {/* <MenuItem value="contractservice_getCurrentHiveCyle">Get Current Hive Cycle</MenuItem>
                          <Divider />
                          <MenuItem value="contractservice_registerOperator">Register Operator</MenuItem>
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

export default RenderRequestPlayground;