import * as yup from 'yup'
import { useEffect, useState } from "react";
import FormContainer from '../../components/form/FormContainer';
import InputField from "../../components/form/InputField";
import MultiStepForm, { FormStep } from "../../components/form/MultistepForm";
import { Box, Button, Grid, Link, Typography } from "@mui/material";
import { Stack } from "@mui/system";

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelectionDialog } from '../../components/wallets/WalletSelectionDialog';

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";
import { TRUE } from 'sass';

// Validation schema for email adresses
const validationSchema_PersonForm = yup.object({
  username: yup.string().required('Username is required'),
  email: yup.string().email('A valid email in the form "mail@domain.de" address is required').required('Email is required')
})

const SignUp = () => {
  const [open, setOpen] = useState(false);
  const { accountId, walletInterface } = useWalletInterface();

  const handleConnect = async () => {
    if (accountId) {
      walletInterface.disconnect();
    } else {
      setOpen(true);
    }
  };

  useEffect(() => {
    if (accountId) {
      setOpen(false);
    }
  }, [accountId])

  return (
    <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="75vh">
      <Typography variant="h5" color="text.primary">
        Sign Up
      </Typography>

      {/* Sign up form */}
      <FormContainer>
        <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo) ' className='renderhiveLogoImg' style={{width: 48, marginBottom: 10}}/>

        {/* Multistep form for the sign-up process */}
        {/* TODO: Here we will later need something that passes the information to the backend */}
        <MultiStepForm
          initialValues={{
            username: '',
            email: '',
            accountID: {accountId},
            node_name: '',
            node_allowance: 1,
          }}
          onSubmit={(values) => {
            alert(JSON.stringify(values, null, 2))
          }}
          showStepper={true}
        >
          <FormStep stepName="Node Operator Details" onSubmit={() => console.log('Step1 submit')} validationSchema={validationSchema_PersonForm}>
            <InputField name="username" label="Username"/>
            <InputField name="email" label="Email"/>
            <Grid container spacing={0} alignItems="flex-end">
                <Grid item xs>
                    <InputField disabled name="accountid" label="AccountID" value={accountId || ''} />
                </Grid>
                <Grid item>
                    <Button
                    variant='contained'
                    sx={{
                      ml: "auto",
                      height: "57px",
                    }}
                    onClick={handleConnect}
                  >
                    {accountId ? `Disconnect` : 'Connect Wallet'}
                  </Button>
                </Grid>
            </Grid>
            <Box display="flex" justifyContent="space-between">

            </Box>
          </FormStep>

          <FormStep stepName="Create Node Account" onSubmit={() => console.log('Step3 submit')}>
            <InputField name="node_name" label="Node Name"/>
            <InputField name="node_allowance" label="Node Funding"/>
          </FormStep>

          <FormStep stepName="Verify" onSubmit={() => console.log('Step4 submit')}>
          
          </FormStep>
        </MultiStepForm>
        
        <Typography fontSize={12}>
          <Box marginTop="20px">Already registered your Renderhive account? <Link href="/signin">Sign in!</Link></Box>
        </Typography>
      </FormContainer>

      {/* Overlay for wallet selection */}
      <WalletSelectionDialog open={open} onClose={() => setOpen(false)} />

    </Box>
  )
}

export default SignUp;