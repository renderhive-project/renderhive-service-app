import { AppBar, Avatar, Badge, Box, Button, Divider, IconButton, Toolbar, Tooltip, Typography, useTheme } from '@mui/material';

// icons
import NotificationsIcon from '@mui/icons-material/Notifications';
import SettingsIcon from '@mui/icons-material/Settings';

// images
import RenderhiveLogo from "../../assets/renderhive-logo.svg";

// styles
import "./navbar.scss"

const NavBar = () => {

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

        <Box sx={{ display: { xs: 'none', md: 'flex' } }}>
          <IconButton
            color="inherit"
            sx={{ borderRadius: '0%', padding: '16px' }}
          >
            <Badge badgeContent={0} color="error">
              <NotificationsIcon />
            </Badge>
          </IconButton>
          <IconButton
            sx={{ borderRadius: '0%', padding: '12px' }}
          >
            <Avatar alt="Node Operator" src="/static/images/avatar/2.jpg" />
          </IconButton>

          <IconButton
            edge="end"
            color="inherit"
            sx={{ borderRadius: '0%', padding: '16px' }}
          >
            <SettingsIcon />
          </IconButton>
        </Box>
      </Toolbar>
      <Divider />
    </AppBar>

  )
}

export default NavBar