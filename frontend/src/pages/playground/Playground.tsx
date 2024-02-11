import { appConfig } from "../../config";
import { useState, useEffect } from 'react';
import { useSession } from '../../contexts/SessionContext';
import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// components
import { Box, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material'

// styles
import "./playground.scss"
import SmartContractPlayground from "./SmartContractPlayground";
import RenderRequestPlayground from "./RenderRequestPlayground";
import RenderOfferPlayground from "./RenderOfferPlayground";

// define playgrounds
const playgrounds = [
  {
    value: 'smart_contract',
    name: 'Smart Contract Playground',
    component: <SmartContractPlayground />
  },
  {
    value: 'render_request',
    name: 'Render Request Playground',
    component: <RenderRequestPlayground />
  },
  {
    value: 'render_offer',
    name: 'Render Offer Playground',
    component: <RenderOfferPlayground />
  },
];

const Playground = () => {
  const [selectedPlayground, setSelectedPlayground] = useState(playgrounds[0].value);

  const handlePlaygroundChange = (event: SelectChangeEvent<string>) => {
    setSelectedPlayground(event.target.value);
  };

  return (
    <Box className="playground">
      <Box width={{ xs: '100%', sm: '90%', md: '80%', lg: '80%' }}>
        
        {/* Page Title */}
        <Typography variant="h4" sx={{ marginBottom: '25px', }}>JSON-RPC Playground</Typography>

        <Select
          value={selectedPlayground}
          onChange={handlePlaygroundChange}
          sx={{ marginBottom: '10px' }}
          fullWidth
        >
          {playgrounds.map((playground) => (
            <MenuItem key={playground.value} value={playground.value}>
              {playground.name}
            </MenuItem>
          ))}
        </Select>

        {playgrounds.map((playground) => {
          if (playground.value === selectedPlayground) {
            return playground.component;
          }
          return null;
        })}

      </Box>
    </Box>
  );
}

export default Playground;