
import { appConfig } from "../../config";
import { useContext, useEffect, useState } from 'react';
import { useSession } from '../../contexts/SessionContext';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// components
import FormContainer from '../../components/form/FormContainer';
import { Box, Button, CircularProgress, Divider, Link, TextField, Typography } from "@mui/material";
import Stack from '@mui/material/Stack';

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelector } from '../../components/wallets/SignInWalletSelector';
import { connectToBladeWallet } from "../../services/wallets/blade/bladeClient";
import { hashConnect } from "../../services/wallets/hashconnect/hashconnectClient";
import { connectToMetamask } from "../../services/wallets/metamask/metamaskClient";

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";


const SignIn = () => {
  const [loading, setLoading] = useState(false);
  const { accountId, walletInterface } = useWalletInterface();
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const { signedIn, setSignedIn, operatorInfo, setOperatorInfo, nodeInfo, setNodeInfo } = useSession();

  // function to handle the login
  const handleSignIn = async () => {
    try {

        setLoading(true);
        setErrorMessage("");
        setStatusMessage("Preparing sign in data ...");

        // Request payload from backend in a first RPC request
        const response_payload = await axios.post(appConfig.api.jsonrpc.url, {
          jsonrpc: '2.0',
          method: 'OperatorService.GetSignInPayload',
          params: [{}],
          id: uuidv4()
        }, {
          headers: {
              'Content-Type': 'application/json',
          },
          withCredentials: true, // This will include cookies in the request
        });

        // Request signing of the payload from the wallet app
        if (response_payload.data && response_payload.data.result) {

            // status update
            setStatusMessage("Waiting for user verification ...");

            // Grab the topic and account to sign from the last pairing event
            const pairingData = hashConnect.hcData.pairingData[hashConnect.hcData.pairingData.length - 1];

            // request signing of the data
            const response_signature = await hashConnect.sign(pairingData.topic, pairingData.accountIds[0], response_payload.data.result.Payload);
            if (response_signature.success) {

                // status update
                setStatusMessage("Signing in and loading data ...");

                // obtain the Uint8Array representations for submission to the backend
                let userSignature = response_signature.userSignature as Uint8Array;
                let signedPayload = new Uint8Array(Buffer.from(JSON.stringify(response_signature.signedPayload)))

                // Send the signed payload to the backend for signing in
                const response_signin = await axios.post(appConfig.api.jsonrpc.url, {
                  jsonrpc: '2.0',
                  method: 'OperatorService.SignIn',
                  params: [{
                      UserSignature: Array.from(userSignature),
                      SignedPayload: Array.from(signedPayload)
                  }],
                  id: uuidv4()
                }, {
                  headers: {
                      'Content-Type': 'application/json',
                  },
                  withCredentials: true, // This will include cookies in the request
                });

                // if successfully signed in
                if (response_signin.data && response_signin.data.result) {
                    
                    // status update
                    setStatusMessage("Signed in sucessfully!");

                    // update status
                    setLoading(false);
                    setErrorMessage("");
                    setStatusMessage("");
                    setSignedIn(response_signin.data.result.SignedIn)

                    // TODO: HANDLE SESSION CONTEXT
                    // AFTER THE USER WAS VERIFIED BY ITS SIGNATURE, WE NEED A SESSION
                    // HANDLING
                    //    - create a session token
                    //    - create a SessionClient component that checks if the user is in an active session using the token and the backend
                    //    - look at the HashconnectClient instance for example

                } else {
                    console.log()
                    setLoading(false);
                    setStatusMessage("");
                    setErrorMessage(`Error: ${response_signin.data.error.message}. Are you connected to the correct wallet?`);
                    console.error('Error signing in:', response_signin.data.error.message);
                }

            } else {
              setLoading(false);
              setStatusMessage("");
              setErrorMessage(`Error signing in: '${response_signature.error}'. Please try again.`);
              console.error('Error signing in:', response_signature.error);
            }

        } else {
            setLoading(false);
            setStatusMessage("");
            setErrorMessage('Unexpected server response.');
            console.error("Unexpected response format:", response_payload.data);
        }

        
    } catch (error) {
        setLoading(false);
        setStatusMessage("");
        setErrorMessage(`Error signing in (${error}). Please try again.`);
        console.error('Error signing in:', error);
    }

    return signedIn;
  };

  // return page content
  return (
    <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="80vh">

        {/* Sign in form */}
        <FormContainer>
          <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo) ' className='renderhiveLogoImg' style={{width: 48}}/>

          <Stack display="flex" flexDirection="column" width="100%" p={2} gap={1}>

              <Typography variant="h6" color="text.primary">
                <b>.: Welcome{accountId && `, ${operatorInfo.username}!`} :.</b>
              </Typography>
              <Typography fontSize={12} marginBottom={'20px'}>
                {
                  !accountId 
                    ? "Please connect to your registered Renderhive user wallet:" 
                    : (
                      <>
                        {`You are connected to your Renderhive user wallet ${accountId}.`}
                      </>
                    )
                }
              </Typography>


              {loading ? (
                <Box justifyContent="center">
                  <CircularProgress />
                  <Typography variant="body2">{statusMessage}</Typography>
                </Box>
              ) : (
                <>
                  { !accountId ? (
                    <WalletSelector/>
                  ) : (
                    <>
                      <TextField disabled name="node_alias" label="Sign in to your Renderhive node:" value={nodeInfo.alias}/>
                      <Button
                        fullWidth
                        variant="outlined"
                        onClick={async () => {
                          await handleSignIn();
                        }}
                      >
                        Sign In
                        
                      </Button>
                  
                    </>
                  )}
                  
    
                  {errorMessage && 
                    <Box justifyContent="center" justifyItems="center">
                      {/* <ErrorOutlineIcon color="error"/>  */}
                      <Typography color="error" variant="body2">{errorMessage}</Typography>
                    </Box>
                  }
                  
                  <Divider>
                    <Typography variant="h6" color="text.primary">or</Typography>
                  </Divider>
    
                  {accountId && 
                    <>
                      <Button
                        fullWidth
                        disabled={!accountId}
                        variant='contained'
                        sx={{
                          ml: "auto"
                        }}
                        onClick={() => {
                          walletInterface.disconnect()
                        }}
                      >
                        Disconnect Wallet
                      </Button>
                    </>
                  }
    
                  <Button fullWidth href="/signup" variant="contained" color="primary">
                    SIGN UP
                  </Button>
    
                  <Box marginTop="20px">
                    <Typography fontSize={12}>
                      Never used Renderhive before? <br></br>
                      Check out the <Link href="/signup">Get Started Guide</Link>!
                    </Typography>
                  </Box>
                </>
              )}

          </Stack>
        </FormContainer>
    </Box>
  )
}

export default SignIn;