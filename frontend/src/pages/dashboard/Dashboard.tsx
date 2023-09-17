import React from 'react'
import { Box, Grid, Typography, useTheme } from '@mui/material'

// styles
import "./dashboard.scss"
import { tokens } from '../../theme'
import BasicCard from '../../components/cards/BasicCard'

const Dashboard = () => {
  const theme = useTheme();
  const colors = tokens(theme.palette.mode);

  return (
    <Box className="dashboard">

      {/* Box 1 */}
      <Box width={{ xs: '100%', sm: '90%', md: '80%', lg: '80%' }}>

        {/* // Page Title */}
        <Typography variant="h4" sx={{ marginBottom: '10px', }}>Dashboard</Typography>

        {/* // Page Content */}
        <Grid container spacing={0} className="page-content" bgcolor="background.paper">
          <Grid item xs={12} >
    
            <Grid container spacing={2} direction="row">
              <Grid item xs={6}>
                <Grid container spacing={3}>
                  <Grid item xs={4}>
                    <BasicCard type="cardAccountBalance" title="Total Balance" icon=""/>
                  </Grid>
                  <Grid item xs={4}>
                    <BasicCard type="cardAccountBalance" title="Total Revenue" icon=""/>
                  </Grid>
                  <Grid item xs={4}>
                    <BasicCard type="cardAccountBalance" title="Total Expenses" icon=""/>
                  </Grid>
                </Grid>
              </Grid>
              <Grid item xs={6}>
                <Grid style={{ height: "100%" }}>
                  <BasicCard type="cardAccountBalance" title="Other elements" icon=""/>
                </Grid>
              </Grid>
            </Grid>

          </Grid>
        </Grid>
        
        <Grid container spacing={0} className="page-content" bgcolor="background.paper">
          <Grid item xs={12} sx={{height: '500px'}}>
    

          </Grid>
        </Grid>

      </Box>

    </Box>

  )
}

export default Dashboard