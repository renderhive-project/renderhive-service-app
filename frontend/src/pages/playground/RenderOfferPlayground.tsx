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
  renderhiveservice_createrenderoffer: `{
    "jsonrpc": "2.0",
    "method": "NodeService.CreateRenderOffer",
    "params": {
      "blenderversions": [
        {
          "version": "4.0.2",
          "engines": ["CYCLES"],
          "devices": ["METAL", "CUDA", "OPTIX"],
          "threads": 4
        }
      ],
      "price": 0.1
    }
  }`,
  renderhiveservice_submitrenderoffer: `{
    "jsonrpc": "2.0",
    "method": "NodeService.SubmitRenderOffer",
    "params": {
      "RenderOfferCID": ""
    }
  }`,
  renderhiveservice_pauserenderoffer: `{
    "jsonrpc": "2.0",
    "method": "NodeService.PauseRenderOffer",
    "params": {
      "RenderOfferCID": ""
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

const RenderOfferPlayground = () => {
    const [loading, setLoading] = useState(false);
    const { accountId, walletInterface } = useWalletInterface();
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [successMessage, setSuccessMessage] = useState<string>('');
    const [editorText, setEditorText] = useState<string>('');
  
    // set the default template
    const defaultTemplateKey = "renderhiveservice_createrenderoffer";
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
  
    // send a offer to the backend (JSON-RPC)
    const sendJsonRpcOffer = async (offerBody: object) => {
      try {
        setLoading(true);
        setSuccessMessage("");
    
        // add headers, credentials, and send the offer
        const response = await axios.post(appConfig.api.jsonrpc.url, offerBody, {
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
  
    // execute the offer
    const executeOffer = async () => {
  
      // define the offers that are processed by the frontend
      const frontend_offers = ['NodeService.SubmitRenderOffer', 'NodeService.PauseRenderOffer'];
  
      // get the offer body from the editor
      let offerBody = JSON.parse(editorText);
  
      // if this is the create render offer command
      if (offerBody.method === "NodeService.CreateRenderOffer") {

          // serialize the files to base64 and add them to the offer body
          const files = await Promise.all(
            offerFiles.map((file) => 
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

          // Add the files array to the params field of the offer body
          offerBody.params = {
            ...offerBody.params,
            files: files,
          };

      }
      
      // Add the id and timeout fields to the offer body
      offerBody = {
        ...offerBody,
        id: uuidv4(),
        timeout: appConfig.constants.BACKEND_JSONRPC_TIMEOUT,
      };

      console.log(offerBody);

      // if the offer method is NOT in the frontend_offers array
      if (!frontend_offers.includes(offerBody.method)) {
          // send the offer to the backend
          return sendJsonRpcOffer(offerBody);
  
      // execute the offer in the frontend
      } else {
  
          try {
  
            // set loading and success message
            setLoading(true);
            setSuccessMessage("");
  
            // add headers, credentials, and send the offer
            const response = await axios.post(appConfig.api.jsonrpc.url, offerBody, {
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
    const [offerFiles, setOfferFiles] = useState<File[]>([]);
    const { getRootProps, getInputProps } = useDropzone({
      onDrop: (acceptedFiles) => {
        setOfferFiles(acceptedFiles);
      },
    });

    const fileTable = offerFiles.length > 0 && (
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
              {offerFiles.map((file) => (
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
            {/* Render Offer Playground */}
            <Grid container spacing={0} className="page-content" bgcolor="background.paper">

                {/* Title */}
                <Typography variant="h5" sx={{ marginBottom: '10px', }}>Render Offers</Typography>
                
                <Grid item xs={12}>
                  <Grid container spacing={4}>
                    <Grid item xs={9}>
                      <AceEditor
                          mode="json"
                          theme="github_dark"
                          onChange={setEditorText}
                          name="rpc-offer-editor"
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

                          {/* Select Offer Template */}
                          <Select
                          defaultValue={defaultTemplateKey}
                          onChange={handleTemplateChange}
                          fullWidth
                          >
                          <MenuItem value="renderhiveservice_createrenderoffer">Create Render Offer</MenuItem>
                          <MenuItem value="renderhiveservice_submitrenderoffer">Submit Render Offer</MenuItem>
                          <MenuItem value="renderhiveservice_pauserenderoffer">Pause Render Offer</MenuItem>
                          {/* <MenuItem value="contractservice_getCurrentHiveCyle">Get Current Hive Cycle</MenuItem>
                          <Divider />
                          <MenuItem value="contractservice_registerOperator">Register Operator</MenuItem>
                          {/* Add more menu items as needed */}
                          </Select>
                      </Box>
                      
                      {/* Send Offer */}
                      <LoadingButton
                          fullWidth
                          onClick={executeOffer}
                          setError={setErrorMessage}
                          loadingText="Sending ..."
                          buttonText="Send Offer"
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

export default RenderOfferPlayground;