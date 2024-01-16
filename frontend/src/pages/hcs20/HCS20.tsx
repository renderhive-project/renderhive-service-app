import React, { useState, useContext, Fragment } from 'react';
import { Button,FormControlLabel, Switch, CircularProgress, Accordion, AccordionSummary, AccordionDetails, Typography, TextField, Container, Paper, TableContainer, Table, TableBody, TableRow, TableHead, TableCell } from '@mui/material';
import axios from 'axios';

const HCS20 = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [topicId, setTopicId] = useState('0.0.4350190');
  const [balances, setBalances] = useState(null);
  const [hasSubmitKey, setHasSubmitKey] = useState(false); // New state for submit key toggle

  const displayBalances = async () => {
    if (!topicId) {
      alert('Please enter a topic ID.');
      return;
    }
    setIsLoading(true);
    const fetchedBalances = await getHcs20DataFromTopic(topicId);
    console.log('fetchedBalances', fetchedBalances);
    setBalances(fetchedBalances);
    setIsLoading(false);
  };

  async function getHcs20DataFromTopic(
    topicId: string, 
    allMessages: any[] = [], 
    invalidMessages: string[] = [], 
    nextLink?: string
  ) {
    const baseUrl = 'https://mainnet-public.mirrornode.hedera.com';
    const url = nextLink ?  `${baseUrl}${nextLink}` : `${baseUrl}/api/v1/topics/${topicId}/messages?limit=1000&order=asc`;

    try {
      const response = await axios.get(url);
      const { messages, links } = response.data;
    
      messages.forEach(msg => {
        try {
          const parsedMessage = JSON.parse(atob(msg.message));
          parsedMessage.payer_account_id = msg.payer_account_id
          allMessages.push(parsedMessage);
        } catch (error) {
          console.error('Error parsing message:', error);
          invalidMessages.push(msg.consensusTimestamp); // Tracking invalid message
        }
      });

      if (links && links.next && allMessages.length < 10000) {
        console.log(topicId, allMessages, invalidMessages, links.next);
        return await getHcs20DataFromTopic(topicId, allMessages, invalidMessages, links.next);
      }
      return { balances: await calculateBalances(allMessages, topicId), invalidMessages };

    } catch (error) {
      console.error('Error fetching topic data:', error);
    }
  }


  async function calculateBalances(messages, topicId) {
    const balances = {};
    const transactionsByAccount = {};
    const failedTransactions = [];
    const tokenConstraints = {}; // Store max and lim constraints for each token
    const tokenDetails = {};

    const requiresMatchingPayer = !hasSubmitKey;

    messages.forEach(({ op, tick, amt, from, to, payer_account_id, max, lim, metadata, m }) => {
      // Initialize token constraints and balances
      if (!balances[tick]) {
        balances[tick] = {};
        tokenConstraints[tick] = { max: Infinity, lim: Infinity, totalMinted: 0 };
      }
      if (!transactionsByAccount[tick]) {
        transactionsByAccount[tick] = {};
      }

      const amount = parseInt(amt);
      let failureReason = '';

      switch (op) {
        case 'deploy':
          // Set max and lim constraints for the token
          tokenConstraints[tick].max = max ? parseInt(max) : Infinity;
          tokenConstraints[tick].lim = lim ? parseInt(lim) : Infinity;

          tokenDetails[tick] = {
            maxSupply: parseInt(max) || 'Not Set',
            currentSupply: 0, // Initialize current supply
            lim: parseInt(lim) || 'Not Set',
            metadata: metadata || 'No Metadata',
            memo: m || '',
          };
          break;
        case 'mint':

          if (tokenDetails[tick]) {
            tokenDetails[tick].currentSupply += parseInt(amt);
          } else  {
            failureReason = 'No Deploy transaction for this Mint tick';
            failedTransactions.push({ op, tick, amt, to, payer_account_id, failureReason });
          }

          if (amount > tokenConstraints[tick].lim) {
            failureReason = 'Mint amount exceeds limit per transaction.';
            failedTransactions.push({ op, tick, amt, to, payer_account_id, failureReason });
            return;
          }
          if (tokenConstraints[tick].totalMinted + amount > tokenConstraints[tick].max) {
            failureReason = 'Mint amount exceeds maximum supply.';
            failedTransactions.push({ op, tick, amt, to, payer_account_id, failureReason });
            return;
          }
          tokenConstraints[tick].totalMinted += amount;
          balances[tick][to] = (balances[tick][to] || 0) + amount;
          break;
        case 'burn':
          if (balances[tick][from] >= amount && (!requiresMatchingPayer || payer_account_id === from)) {
            balances[tick][from] -= amount;
          } else {
            failureReason = balances[tick][from] < amount 
              ? 'Insufficient balance for burn operation.'
              : 'Payer account ID does not match the account from which points are being burned.';
            failedTransactions.push({ op, tick, amt, from, to, payer_account_id, failureReason });
            return;
          }
          break;
        case 'transfer':
          if (balances[tick][from] >= amount && (!requiresMatchingPayer || payer_account_id === from)) {
            balances[tick][from] -= amount;
            balances[tick][to] = (balances[tick][to] || 0) + amount;
          } else {
            failureReason = balances[tick][from] < amount 
              ? 'Insufficient balance for transfer operation.'
              : 'Payer account ID does not match the sender\'s account.';
            failedTransactions.push({ op, tick, amt, from, to, payer_account_id, failureReason });
            return;
          }
          break;
      }

      // Record the transaction
      if (!transactionsByAccount[tick][from]) {
        transactionsByAccount[tick][from] = [];
      }
      if (!transactionsByAccount[tick][to]) {
        transactionsByAccount[tick][to] = [];
      }
      transactionsByAccount[tick][from].push({ op, amt, to, from, m });
      transactionsByAccount[tick][to].push({ op, amt, to, from, m });
    });

    console.log('balances',balances);
    return { balances, transactionsByAccount, failedTransactions, tokenDetails };
  }




  return (
    <React.Fragment>
      <Container>
        <br /> 
        <Typography variant="h4" gutterBottom>
          HCS-20 Balances Viewer
        </Typography>
        <TextField
          label="Enter Topic ID"
          value={topicId}
          onChange={(e) => setTopicId(e.target.value)}
          fullWidth
          margin="normal"
        />
        <br />
        <br />
        <FormControlLabel
        sx={{
          display: 'block',
        }}
        control={
          <Switch
            checked={hasSubmitKey}
            onChange={() => setHasSubmitKey(!hasSubmitKey)}
            name="loading"
            color="primary"
          />
        }
        label="Submit Key"
      />
        <br />
        <br />
        {isLoading ? (
          <>
          <Button variant="contained" color="primary" disabled>
            <CircularProgress size={24} />
          </Button>
          <br />
          <br />
          Indexing...
          </>
        ) : (
          <Button variant="contained" color="primary" onClick={displayBalances}>
            Get Balances
          </Button>
        )}
        {balances && balances.balances.tokenDetails && (
          <Fragment>
          <br />
          <br />
            <Typography variant="h4" gutterBottom style={{ marginTop: '20px' }}>
              Details
            </Typography>
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Tick</TableCell>
                    <TableCell>Max Supply</TableCell>
                    <TableCell>Current Supply</TableCell>
                    <TableCell>Limit per Mint Transaction</TableCell>
                    <TableCell>Metadata</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {Object.entries(balances.balances.tokenDetails).map(([token, details]:[any, any]) => (
                    <TableRow key={token}>
                      <TableCell>{token}</TableCell>
                      <TableCell>{details && details.maxSupply}</TableCell>
                      <TableCell>{details && details.currentSupply}</TableCell>
                      <TableCell>{details && details.lim}</TableCell>
                      <TableCell>{details && details.metadata}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Fragment>
        )}
        <br />
        <br />
        <Typography variant="h4" gutterBottom style={{ marginTop: '20px' }}>
            Balances
          </Typography>
        {balances && (
        <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Account ID</TableCell>
              <TableCell>Tick</TableCell>
              <TableCell align="right">Balance</TableCell>
              <TableCell align="right">Transactions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
        {Object.entries(balances.balances.balances).map(([token, accounts]) => (
          <Fragment key={token}>
            {Object.entries(accounts).map(([accountId, balance]) => (
              <TableRow key={accountId}>
                <TableCell>{accountId}</TableCell>
                <TableCell>{token}</TableCell>
                <TableCell align="right">{balance}</TableCell>
                <TableCell align="right">
                <Accordion>
                    <AccordionSummary>
                      <Typography>View Transactions</Typography>
                    </AccordionSummary>
                    <AccordionDetails style={{ backgroundColor: "#010101" }}>
                      <TableContainer>
                        <Table size="small">
                          <TableHead>
                            <TableRow>
                              <TableCell>Operation</TableCell>
                              <TableCell>Amount</TableCell>
                              <TableCell>From</TableCell>
                              <TableCell>To</TableCell>
                            </TableRow>
                          </TableHead>
                          <TableBody>
                            {balances.balances.transactionsByAccount[token][accountId].map((tx, index) => (
                              <TableRow key={index}>
                                <TableCell>{tx.op}</TableCell>
                                <TableCell>{tx.amt}</TableCell>
                                <TableCell>{tx.from}</TableCell>
                                <TableCell>{tx.to}</TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </TableContainer>
                    </AccordionDetails>
                  </Accordion>
                </TableCell>
              </TableRow>
            ))}
          </Fragment>
        ))}
      </TableBody>
        </Table>
      </TableContainer>
      )}{balances && balances.balances.failedTransactions && balances.balances.failedTransactions.length > 0 && (
        <Fragment>
        <br />
        <br />
          <Typography variant="h4" gutterBottom style={{ marginTop: '20px' }}>
            Failed Transactions
          </Typography>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Operation</TableCell>
                  <TableCell>Tick</TableCell>
                  <TableCell>Amount</TableCell>
                  <TableCell>From</TableCell>
                  <TableCell>To</TableCell>
                  <TableCell>Payer Account ID</TableCell>
                  <TableCell>Reason</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {balances.balances.failedTransactions.map((tx, index) => (
                  <TableRow key={index}>
                    <TableCell>{tx.op}</TableCell>
                    <TableCell>{tx.tick}</TableCell>
                    <TableCell>{tx.amt}</TableCell>
                    <TableCell>{tx.from}</TableCell>
                    <TableCell>{tx.to}</TableCell>
                    <TableCell>{tx.payer_account_id}</TableCell>
                    <TableCell>{tx.failureReason}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Fragment>
      )}
      
        {/* {balances && (
          <Typography variant="body1" gutterBottom>
            <pre>{JSON.stringify(balances, null, 2)}</pre>
          </Typography>
        )} */}
      </Container>
    </React.Fragment>
  );
}

export default HCS20;