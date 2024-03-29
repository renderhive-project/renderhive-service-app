import React, { ReactNode } from 'react';
import {Box} from '@mui/material';

interface FormContainerProps {
  children: ReactNode;
  width?: string;
  height?: string;
}

const FormContainer: React.FC<FormContainerProps> = ({ children, width = "25%", height }) => {

  return (
    <Box
      textAlign="center" 
      sx={{
          backgroundColor: 'background.paper',
          borderRadius: '10px',
          //boxShadow: '0px 2px 5px 0px rgba(0,0,0,0.25)',
          border: '1px solid #222222',
          padding: '20px',
          width: {width},
          height: {height},
          margin: '20px auto'
      }}
    >
      {children}
    </Box>
  );
};

export default FormContainer;