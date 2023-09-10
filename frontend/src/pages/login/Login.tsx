import * as yup from 'yup'
import { useEffect, useState } from "react";
import FormContainer from '../../components/form/FormContainer';
import InputField from "../../components/form/InputField";
import MultiStepForm, { FormStep } from "../../components/form/MultistepForm";
import { Button, Typography } from "@mui/material";
import { Stack } from "@mui/system";

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelectionDialog } from '../../components/wallets/WalletSelectionDialog';

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";

// Validation schema for email adresses
const validationSchema_PersonForm = yup.object({
  username: yup.string().required('Username is required'),
  email: yup.string().email('A valid email in the form "mail@domain.de" address is required').required('Email is required')
})

export default function SignUp() {
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
    <Stack alignItems="center" spacing={4}>
      <Typography variant="h4" color="white">
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
            node_alias: '',
            node_allowance: 1,
          }}
          onSubmit={(values) => {
            alert(JSON.stringify(values, null, 2))
          }}
        >
          <FormStep stepName="Node Operator Details" onSubmit={() => console.log('Step1 submit')} validationSchema={validationSchema_PersonForm}>
            <InputField name="username" label="Username"/>
            <InputField name="email" label="Email"/>
          </FormStep>

          <FormStep stepName="Connect Operator Wallet" onSubmit={() => console.log('Step2 submit')}>
            <Button
              variant='contained'
              sx={{
                ml: "auto"
              }}
              onClick={handleConnect}
            >
              {accountId ? `Connected: ${accountId}` : 'Connect Operator Wallet'}
            </Button>
          </FormStep>

          <FormStep stepName="Create Node Account" onSubmit={() => console.log('Step3 submit')}>
            <InputField name="node_ailas" label="Node Alias"/>
            <InputField name="node_allowance" label="Node Funding"/>
          </FormStep>

          <FormStep stepName="Verify" onSubmit={() => console.log('Step4 submit')}>
          
          </FormStep>
        </MultiStepForm>

      </FormContainer>

      {/* Overlay for wallet selection */}
      <WalletSelectionDialog open={open} onClose={() => setOpen(false)} />

    </Stack>
  )
}