import { appConfig } from "../../config";
import { useState } from 'react';
import { useSession } from '../../contexts/SessionContext';
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import axios from 'axios';

// components
import { AppBar, Badge, Box, Button, Divider, IconButton, Menu, MenuItem, Toolbar, Typography } from '@mui/material';

// icons
import NotificationsIcon from '@mui/icons-material/Notifications';
import SettingsIcon from '@mui/icons-material/Settings';

// images
import RenderhiveLogo from "../../assets/renderhive-logo.svg";

// styles
import "./navbar.scss"

const NavBar = () => {
  const [userMenu, setUserMenu] = useState(null);
  const { accountId, walletInterface } = useWalletInterface();
  const navigate = useNavigate();
  const { signedIn, setSignedIn } = useSession();

  const handleMenuOpen = (event: any) => {
    setUserMenu(event.currentTarget);
  };
  
  const handleMenuClose = () => {
    setUserMenu(null);
  };

  // TODO: Implement a OperatorService.SignOut method on the JSON-RPC, which informs the
  //       backend about the sign out process
  const handleMenuLogout = async () => {
    try {

        // Request payload from backend in a first RPC request
        const response_signout = await axios.post(appConfig.api.jsonrpc.url, {
          jsonrpc: '2.0',
          method: 'OperatorService.SignOut',
          params: [{}],
          id: uuidv4()
        }, {
          headers: {
              'Content-Type': 'application/json',
          },
          withCredentials: true, // This will include cookies in the request
        });

        // if signout succeeded
        if (response_signout.data && response_signout.data.result) {

            // sign out
            setSignedIn(false)
            console.log(response_signout.data.result.Message);

        } else {
            console.error("Unexpected response format:", response_signout.data);
        }
        
    } catch (error) {
        console.error('Error signing out:', error);
    }

    if (walletInterface) {
        // OPTIONAL: disconnect from the operator wallet?
        // walletInterface.disconnect();
    }

    // close the menu
    handleMenuClose();
    navigate("/signin");
  };

  return (
    <AppBar position="fixed" sx={{ justifyContent: 'center', align: 'center', zIndex: (theme) => theme.zIndex.drawer + 1 }} elevation={0}>
      <Toolbar disableGutters={true} sx={{ paddingLeft: '18px', paddingRight: '18px' }}>
        <Box sx={{ display: 'flex' }}>
          <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo)' className='renderhiveLogoImg' />
          <Typography variant="h6" pl={2} mt={0.25} sx={{ display: { xs: 'none', md: 'flex' } }} noWrap>
            Renderhive Service App
          </Typography>
        </Box>
        <Box sx={{ flexGrow: 1 }} />

        {/* if a wallet is connected and the account signed in*/}
        {((accountId && signedIn) ?
          <>
            <Box sx={{ display: { xs: 'none', md: 'flex' } }}>
              <IconButton
                color="inherit"
                sx={{ borderRadius: '0%', padding: '16px' }}
              >
                <Badge badgeContent={0} color="error">
                  <NotificationsIcon />
                </Badge>
              </IconButton>
              {/* <IconButton
                sx={{ borderRadius: '0%', padding: '12px' }}
              >
                <Avatar alt="Node Operator" src="/static/images/avatar/2.jpg" />
              </IconButton> */}

              <IconButton
                edge="end"
                color="inherit"
                sx={{ borderRadius: '0%', padding: '16px' }}
                onClick={handleMenuOpen}
              >
                <SettingsIcon />
              </IconButton>
              
              {/* The basic user menu, which allows actions like logging out */}
              <Menu
                anchorEl={userMenu}
                open={Boolean(userMenu)}
                onClose={handleMenuClose}
              >
                <MenuItem disabled>Operator: {accountId}</MenuItem>
                <MenuItem onClick={() => {
                  handleMenuLogout();
                }}>
                  Logout
                </MenuItem>
              </Menu>
            </Box>
          </>
        : 
          <Button
            disabled={!accountId}
            variant='contained'
            sx={{
              ml: "auto",
            }}
            onClick={() => {
              (walletInterface && walletInterface.disconnect())
            }}
          >
            {(!accountId) ? "Connect Wallet" : `Disconnect ${accountId}`}
          </Button>
        )}
     
      </Toolbar>
      <Divider />
    </AppBar>

  )
}

export default NavBar