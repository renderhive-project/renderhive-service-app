import React from 'react';
import { styled, useTheme } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import { Card, CardContent } from '@mui/material';
import { tokens } from '../../theme';


// define a type based on stylesMap
type CardTypes = keyof typeof stylesMap; 

// Create a styled card component that accepts a cardStyle prop
const StyledCard = styled(Card)(({ theme, cardStyle }) => ({
    ...cardStyle,
}));

// define the props for the component
interface BasicCardProps {
    type: CardTypes;   
    title: string;
    icon: object;
}

const BasicCard: React.FC<BasicCardProps> = ({ type, title }) => {
    const theme = useTheme();
    const colors = tokens(theme.palette.mode);

    const stylesMap = {
        cardAccountBalance: {
            color: theme.palette.secondary.dark,
            backgroundColor: theme.palette.primary.light,
        },
    };

  return (
    <StyledCard cardStyle={stylesMap[type]}>
      <CardContent>
        <Typography sx={{ fontSize: 14 }} gutterBottom>
          {title}
        </Typography>
      </CardContent>
    </StyledCard>
  );
}

export default BasicCard;

