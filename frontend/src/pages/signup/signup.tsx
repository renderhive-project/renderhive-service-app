import { appConfig } from "../../config";
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { useSession } from '../../contexts/SessionContext';

import * as yup from 'yup'
import { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';
import FormContainer from '../../components/form/FormContainer';
import { Field } from 'formik';
import InputField from "../../components/form/InputField";
import MultiStepForm, { FormStep } from "../../components/form/MultistepForm";
import { Alert, AlertTitle, Box, Checkbox, CircularProgress, FormControlLabel, Link, Stack, Typography } from "@mui/material";

// Web3 services & components
import { AccountId } from "@hashgraph/sdk";
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelector } from "../../components/wallets/SignInWalletSelector";
import { KeyringProvider, useKeyring } from '@w3ui/react-keyring';
// import W3Keyring from '../../components/w3up/Keyring';

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";


// Validation schema for email adresses
const validationSchema_PersonForm = yup.object({
  username: yup.string().required('Username is required'),
  email: yup.string().email('A valid email in the form "mail@domain.de" address is required').required('Email is required')
})

const SignUp = () => {
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const { accountId, walletInterface } = useWalletInterface();
  const [{ account }, { loadAgent, unloadAgent, authorize, cancelAuthorize }] = useKeyring()
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const { signedIn, setSignedIn, operatorInfo, setOperatorInfo, nodeInfo, setNodeInfo } = useSession();
  const navigate = useNavigate();

  // sign up process steps
  const steps = [
    {
      label: 'Call Smart Contract',
      description: `Please confirm the transaction to sign up the operator account ${accountId} as node operator in the Renderhive smart contract.`,
    },
    {
      label: 'Sign up this ',
      description:
        'An ad group contains one or more ads which target a shared set of keywords.',
    },
    {
      label: 'Create an ad',
      description: `Try out different ad text to see what brings in the most customers,
                and learn how to enhance your ads using features like ad extensions.
                If you run into any problems with your ads, find out how to tell if
                they're running and how to resolve approval issues.`,
    },
  ];

  // get the connected wallet / Hedera account
  useEffect(() => {
    if (accountId) {
      setOpen(false);
    }
  }, [accountId])

  // load the w3Up agent - once.
  useEffect(() => { loadAgent() }, [])

  // handle form submission
  const handleSubmit = async e => {
    // e.preventDefault()
    // setSubmitted(true)

    // if the wallet interface is initialized
    if (!walletInterface) {
        return
    }

    // alert(JSON.stringify(e, null, 2))
    
    // update status
    setLoading(true);
    setErrorMessage("");

    // status update
    setLoading(true);
    setErrorMessage("");
    setStatusMessage("Initialing sign up procedure ...");

    // prepare operator information
    if ((operatorInfo && accountId) && operatorInfo.accountId != accountId) {
      operatorInfo.accountId = accountId
      operatorInfo.username = e.username
      operatorInfo.email = e.email
    }

    // prepare node information
    if (nodeInfo && e.node_name) {
      nodeInfo.name = e.node_name
    }
    
    // STEP 1: Initialize the sign up procedure
    const response_init = await axios.post(appConfig.api.jsonrpc.url, {
      jsonrpc: '2.0',
      method: 'OperatorService.SignUp',
      params: [{
        Step: 'init',
        Operator: operatorInfo,
        Node: nodeInfo,
        Passphrase: e.node_passphrase,
      }],
      id: uuidv4()
    }, {
      headers: {
          'Content-Type': 'application/json',
      },
      withCredentials: true, 
    });

    // if successfully initialized
    if (response_init.data && response_init.data.result) {

        // status update
        setLoading(true);
        setErrorMessage("");
        setStatusMessage("Waiting for user to sign account creation transaction ...");

        // request signing of the data
        const response_createaccount_txnId = await walletInterface.transferHBAR(AccountId.fromString(response_init.data.result.NodeAccountID), 10)
        if (response_createaccount_txnId) {

            // status update
            setStatusMessage("Creating account ...");

            // STEP 2: Create node account
            const response_create = await axios.post(appConfig.api.jsonrpc.url, {
              jsonrpc: '2.0',
              method: 'OperatorService.SignUp',
              params: [{
                Step: 'create',
                Operator: operatorInfo,
                Node: nodeInfo,
                Passphrase: e.node_passphrase,
                AccountCreationTransactionID: response_createaccount_txnId,
              }],
              id: uuidv4()
            }, {
              headers: {
                  'Content-Type': 'application/json',
              },
              withCredentials: true, 
            });
            
            // if successfully saved
            if (response_create.data && response_create.data.result) {
                
                // the account was sucessfully created, so save the new accountID in the state
                setNodeInfo({
                  name: nodeInfo.name,
                  accountId: response_create.data.result.NodeAccountID,
                })

                // status update
                setStatusMessage("Signing up with the smart contract ... ");

                // TODO: HANDLE FURTHER SIGNUP STEPS
                //    - register with smart contract
                //    - register with w3up

                // TODO: Register machine as w3up agent
                // try {
                //   await authorize(operatorInfo.email)
                // } catch (err) {
                //   throw new Error('failed to authorize', { cause: err })
                // // } finally {
                // //   setSubmitted(false)
                // }
                
                // after successfull sign up, go to sign in page
                setStatusMessage("Redirecting to sign-in ...");

                // Redirect user to the signin page
                navigate("/signin");

                // update status
                setLoading(false);
                setErrorMessage("");
                setStatusMessage("");
                
            } else {
                console.log()
                setLoading(false);
                setStatusMessage("");
                setErrorMessage(`Error: ${response_create.data.error.message}. Are you connected to the correct wallet?`);
                console.error('Error signing in:', response_create.data.error.message);
            }

        } else {
          setLoading(false);
          setStatusMessage("");
          // setErrorMessage(`Error signing in: '${response_init.error}'. Please try again.`);
          // console.error('Error signing in:', response_init.error);
        }

    } else {
        console.log()
        setLoading(false);
        setStatusMessage("");
        setErrorMessage(`Error: ${response_init.data.error.message}. Are you connected to the correct wallet?`);
        console.error('Error signing in:', response_init.data.error.message);
    }

  }

  return (
    <KeyringProvider>
    <Box bgcolor="background.default" display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="75vh" mt="60px">

      <FormContainer width={accountId ? '33%' : undefined}>
        <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo) ' className='renderhiveLogoImg' style={{width: 48}}/>

        <Stack display="flex" flexDirection="column" width="100%" p={2} gap={1}>

            <Typography variant="h6" color="text.primary">
              { (!operatorInfo || (operatorInfo && operatorInfo.accountId != accountId)) && 
                <b>.: Welcome :.</b>
              }
              { ((operatorInfo && operatorInfo.accountId == accountId)) && 
                <b>.: Welcome, {operatorInfo.username}! :.</b>
              }
            </Typography>
            <Typography fontSize={12} marginBottom={'20px'}>
              {
                !accountId 
                  ? "Please connect to the Hedera account you want to use as your user wallet:" 
                  : (
                    <>
                      {`You are connected to your Hedera account ${accountId} and ready to sign up this machine as a new Renderhive node.`}
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
                  <>
                    <WalletSelector/>

                    <Box marginTop="20px">
                      <Typography fontSize={12}>
                        Never used Renderhive before? <br></br>
                        Check out the <Link href="/signup">Get Started Guide</Link>!
                      </Typography>
                    </Box>
                  </>
                ) : (
                  <>

                    {/* Multistep form for the sign-up process */}
                    {/* TODO: Here we will later need something that passes the information to the backend */}
                    <MultiStepForm
                      initialValues={{
                        username: ((operatorInfo && operatorInfo.accountId == accountId) ? operatorInfo.username : ''),
                        email: ((operatorInfo && operatorInfo.accountId == accountId) ? operatorInfo.email : ''),
                        accountID: {accountId},
                        node_name: (nodeInfo ? nodeInfo.name : ''),
                        node_passphrase: '',
                        w3up_signup: false,
                      }}
                      onSubmit={handleSubmit}
                      showStepper={true}
                    >
                      {/* STEP: Register operator account */}
                      { (!operatorInfo || (operatorInfo && operatorInfo.accountId != accountId)) && 
                        <FormStep stepName="Node Operator Details" onSubmit={() => console.log('Step1 submit')} validationSchema={validationSchema_PersonForm}>
                          <Alert severity="info" sx={{ textAlign: 'justify', marginBottom: '20px' }}>
                            <AlertTitle>Info</AlertTitle>
                            The operator details are used to identify your user on the Renderhive network. Your Email
                            address is only stored locally on your machine and not shared without your consent.
                          </Alert>
                          <InputField name="username" label="Username"/>
                          <InputField name="email" label="Email"/>
                        </FormStep>
                      }

                      {/* STEP: Define node details */}
                      <FormStep stepName="Node details" onSubmit={() => console.log('Step2 submit')}>
                        <Alert severity="info" sx={{ textAlign: 'justify', marginBottom: '20px' }}>
                          <AlertTitle>Info</AlertTitle>
                          The node name is used to help you organize your Renderhive nodes and to identify your 
                          node on the Renderhive network.
                        </Alert>
                        <InputField name="node_name" label="Node Name"/>
                        <InputField type="password" name="node_passphrase" label="Node Passphrase"/>
                      </FormStep>

                      {/* STEP: Register with storage service */}
                      <FormStep stepName="Storage space" onSubmit={() => console.log('Step3 submit')}>
                        <Alert severity="warning" sx={{ textAlign: 'justify', marginBottom: '20px' }}>
                          <AlertTitle>Note</AlertTitle>
                          Renderhive stores data connected to render jobs (e.g., Blender files) on the decentralized 
                          Interplanatary File System (IPFS) via a Filecoin service provider. By signing up
                          your Renderhive account, you automatically create a web3.storage / w3up account using your 
                          previously provided email address.
                          <FormControlLabel required 
                            control={
                              <Field 
                                name="w3up_signup"
                                type="checkbox"
                                as={Checkbox}
                              />
                            } 
                            label="Ok, sign me up with web3.storage" 
                            sx={{marginTop: "15px"}}
                          />
                        </Alert>
                        <InputField disabled name="email" label="Operator Email"/>
                      </FormStep>

                      {/* STEP: Start the sign up process */}
                      <FormStep stepName="Verify" onSubmit={() => console.log('Step4 submit')}>
                        <Alert severity="success" sx={{ textAlign: 'justify', marginBottom: '20px' }}>
                          <AlertTitle>Sign up?</AlertTitle>
                          Please check if the provided information is correct! 
                        </Alert>
                        <InputField disabled name="username" label="Username"/>
                        <InputField disabled name="email" label="Email"/>
                        <InputField disabled name="node_name" label="Node Name"/>
                        {/* TODO: List all actions and costs here */}
                      </FormStep>
                    </MultiStepForm>
                
                  </>
                )}

                {errorMessage && 
                  <Box justifyContent="center" justifyItems="center">
                    <Typography color="error" variant="body2">{errorMessage}</Typography>
                  </Box>
                }
              </>
            )}

        </Stack>

      </FormContainer>

      {/* Overlay for wallet selection */}
      {/* <WalletSelectionDialog open={open} onClose={() => setOpen(false)} /> */}

    </Box>
    </KeyringProvider>
  )
}

export default SignUp;