import { Box } from '@mui/material';
import BuiltOnHedera from "../../assets/built-on-hedera.svg";

// styles
import "./footer.scss"

const Footer = () => {

  return (
    <Box className='footer' position='fixed' sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
      <img 
        src={BuiltOnHedera}
        alt='An upper case H with a line through the top and the text Build on Hedera'
        className='builtOnHederaSVG'
      />

      <Box sx={{ flexGrow: 1 }} />

      <Box sx={{ display: { xs: 'none', md: 'flex' } }}>
        2023 © Christian Stolze
      </Box>
          
    </Box>
  );
}

export default Footer