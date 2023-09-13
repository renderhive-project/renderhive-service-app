import { useState } from 'react';
import { useWalletInterface } from '../../services/wallets/useWalletInterface';
import { AppBar, Avatar, Badge, Box, Divider, IconButton, Menu, MenuItem, Toolbar, Typography } from '@mui/material';
import { useNavigate } from 'react-router-dom';

// icons
import NotificationsIcon from '@mui/icons-material/Notifications';
import SettingsIcon from '@mui/icons-material/Settings';

// images
import RenderhiveLogo from "../../assets/renderhive-logo.svg";

// styles
import "./navbar.scss"

const NavBar = () => {
  const [userMenu, setUserMenu] = useState(null);
  const {accountId, walletInterface} = useWalletInterface();
  const navigate = useNavigate();

  const handleMenuOpen = (event) => {
    setUserMenu(event.currentTarget);
  };
  
  const handleMenuClose = () => {
    setUserMenu(null);

    if (walletInterface) {
      walletInterface.disconnect();
      navigate("/signin")
    }
  };

  return (
    <AppBar position="fixed" sx={{ justifyContent: 'center', align: 'center', zIndex: (theme) => theme.zIndex.drawer + 1 }} elevation={0}>
      <Toolbar disableGutters={true} sx={{ paddingLeft: '18px', paddingRight: '18px' }}>
        <Box sx={{ display: 'flex' }}>
          <img src={RenderhiveLogo} alt='A stylized Blender default cube forming a honeycomb (renderhive logo)' className='renderhiveLogoImg' />
          <Typography variant="h6" pl={3} sx={{ display: { xs: 'none', md: 'flex' } }} noWrap>
            Renderhive Service App
          </Typography>
        </Box>
        <Box sx={{ flexGrow: 1 }} />

        {/* if a wallet is connected */}
        {accountId && (
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
              <MenuItem onClick={() => {
                handleMenuClose();
                // Your logout logic here
              }}>
                Logout
              </MenuItem>
            </Menu>
          </Box>
        )}

      </Toolbar>
      <Divider />
    </AppBar>

  )
}

export default NavBar