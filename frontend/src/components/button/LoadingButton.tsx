import React, { useState } from 'react';
import { Box, Button, CircularProgress } from '@mui/material'

interface LoadingButtonProps {
  onClick: () => Promise<any>;
  setError: React.Dispatch<React.SetStateAction<string>>;
  loadingText: string;
  buttonText: string;

  fullWidth?: boolean;
}

const LoadingButton: React.FC<LoadingButtonProps> = ({ onClick, setError, loadingText, buttonText, fullWidth }) => {
  const [loading, setLoading] = useState(false);

  const handleClick = async () => {
    setLoading(true);
    setError('');
    try {
      await onClick();
    } catch (error) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button variant="contained" color="primary" onClick={handleClick} disabled={loading} fullWidth={fullWidth}>
      {loading ? (
        <Box display="flex" alignItems="center">
          <CircularProgress size={14} color="inherit" />
          <Box ml={1}>{loadingText}</Box>
        </Box>
      ) : (
        buttonText
      )}
    </Button>
  );
};

export default LoadingButton;