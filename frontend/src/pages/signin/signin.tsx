import FormContainer from '../../components/form/FormContainer';
import { Box, Button, Divider, Link, Typography } from "@mui/material";
import Stack from '@mui/material/Stack';

// Web3 services
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { WalletSelector } from '../../components/wallets/WalletSelectionDialog';

// images & icons
import RenderhiveLogo from "../../assets/renderhive-logo.svg";

const SignIn = () => {
  // const {accountId, walletInterface} = useWalletInterface();
  // const { isLoading, setLoading } = useLoading();

  // return page content
  return (
    <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="80vh">

        {/* Sign in form */}
        <FormContainer>
          <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo) ' className='renderhiveLogoImg' style={{width: 48}}/>

          <Stack display="flex" flexDirection="column" width="100%" p={2} gap={1}>

              <Typography variant="h6" color="text.primary">
                <b>.: Welcome :.</b>
              </Typography>
              <Typography fontSize={12} marginBottom={'20px'}>
                Please sign in with your favorite Hedera wallet app.
              </Typography>

              <WalletSelector/>
              
              <Divider>
                <Typography variant="h6" color="text.primary">or</Typography>
              </Divider>

              <Button fullWidth href="/signup" variant="contained" color="primary">
                SIGN UP
              </Button>

              <Box marginTop="20px">
                <Typography fontSize={12}>
                  Never used Renderhive before? <br></br>
                  Check out the <Link href="/signup">Get Started Guide</Link>!
                </Typography>
              </Box>
          </Stack>
        </FormContainer>
    </Box>
  )
}

export default SignIn;