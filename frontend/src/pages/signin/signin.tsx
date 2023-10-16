import { appConfig } from "../../config";
import { useState } from 'react';
import { useSession } from '../../contexts/SessionContext';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// components
import FormContainer from '../../components/form/FormContainer';
import { Box, Button, CircularProgress, Link, Typography } from "@mui/material";
import Stack from '@mui/material/Stack';

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelector } from '../../components/wallets/SignInWalletSelector';
// import { connectToBladeWallet } from "../../services/wallets/blade/bladeClient";
// import { hashConnect } from "../../services/wallets/hashconnect/hashconnectClient";
// import { connectToMetamask } from "../../services/wallets/metamask/metamaskClient";

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";
import InputField from "../../components/form/InputField";
import { Form, Formik } from "formik";


const SignIn = () => {
  const [loading, setLoading] = useState(false);
  const { accountId } = useWalletInterface();
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const { signedIn, setSignedIn, operatorInfo, nodeInfo } = useSession();

  // function to handle the login
  const handleSignIn = async (event: any) => {
    try {

        // status update
        setLoading(true);
        setErrorMessage("");
        setStatusMessage("Loading node data ...");

        // Send the signed payload to the backend for signing in
        const response_signin = await axios.post(appConfig.api.jsonrpc.url, {
          jsonrpc: '2.0',
          method: 'OperatorService.SignIn',
          params: [{
              Passphrase: event.node_password,
          }],
          id: uuidv4(),
          timeout: appConfig.constants.BACKEND_JSONRPC_TIMEOUT,
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
    <Box bgcolor="background.default" display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="75vh" mt="60px">

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
                    <Formik
                      initialValues={{
                        node_name: (nodeInfo ? nodeInfo.name : ''),
                        node_password: '',
                      }}
                      onSubmit={handleSignIn}
                    >
                      {() => (
                          <Form>

                            <InputField disabled name="node_name" label="Sign in to your Renderhive node:" value={nodeInfo.name}/>
                            <InputField type="password" name="node_password" label="Node Password"/>
                            
                            <Button
                              fullWidth
                              variant="outlined"
                              type="submit"
                              sx={{ marginTop: "25px" }}
                            >
                              Sign In
                            </Button>
                          </Form>
                        )}
                    </Formik>
                  
                    </>
                  )}
                  
    
                  {errorMessage && 
                    <Box justifyContent="center" justifyItems="center">
                      {/* <ErrorOutlineIcon color="error"/>  */}
                      <Typography color="error" variant="body2">{errorMessage}</Typography>
                    </Box>
                  }
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