import React from 'react';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import { Box, Card, CardContent } from '@mui/material';

const stylesMap = {
    cardAccountBalance: (theme: any) => ({
        color: theme.palette.primary.main,
        backgroundColor: theme.palette.background.paper,
    }),
};

type CardTypes = keyof typeof stylesMap;

const StyledCard = styled(Card)(({ theme, cardStyle }: { theme?: any, cardStyle: CardTypes }) => ({
    ...stylesMap[cardStyle](theme),
}));

interface BasicCardProps {
    type: CardTypes;   
    title: string;
}

const BasicCard: React.FC<BasicCardProps> = ({ type, title }) => {

    return (
        <StyledCard cardStyle={type}>
            <CardContent>
                <Box display="flex" >
                    <Typography sx={{ fontSize: 14 }} alignItems="center" justifyItems="center" gutterBottom>
                        <Box bgcolor="#ffffff" borderRadius={'5px'} width="10px" height="20px" />
                        {title}
                    </Typography>
                </Box>
            </CardContent>
        </StyledCard>
    );
}

export default BasicCard;
