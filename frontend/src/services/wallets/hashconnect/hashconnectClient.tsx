import { HashConnect, HashConnectTypes } from "hashconnect";
import { HashconnectContext } from "../../../contexts/HashconnectContext";
import { useCallback, useContext, useEffect } from 'react';
import { WalletInterface } from "../walletInterface";
import { AccountId, ContractExecuteTransaction, ContractId, Hbar, TokenAssociateTransaction, TokenId, TransferTransaction, Transaction, TransactionResponse } from "@hashgraph/sdk";
import { ContractFunctionParameterBuilder } from "../contractFunctionParameterBuilder";
import { appConfig } from "../../../config";
import { useLoading } from "../../../contexts/LoaderContext";

// icons
import RenderhiveLogo from "../../../assets/renderhive-logo.svg";

const currentNetworkConfig = appConfig.networks.testnet;
const hederaNetwork = currentNetworkConfig.network;

export const hashConnect = new HashConnect();

class HashConnectWallet implements WalletInterface {

  private getSigner() {
    const pairingData = hashConnect.hcData.pairingData[hashConnect.hcData.pairingData.length - 1];
    const provider = hashConnect.getProvider(hederaNetwork, pairingData.topic, pairingData.accountIds[0]);
    return hashConnect.getSigner(provider);
  }

  async transferHBAR(toAddress: AccountId, amount: number) {
    // Grab the topic and account to sign from the last pairing event
    const signer = this.getSigner();

    const transferHBARTransaction = await new TransferTransaction()
      .addHbarTransfer(signer.getAccountId(), -amount)
      .addHbarTransfer(toAddress, amount)
      .freezeWithSigner(signer);

    const txResult = await transferHBARTransaction.executeWithSigner(signer);
    return txResult.transactionId;
  }

  async transferFungibleToken(toAddress: AccountId, tokenId: TokenId, amount: number) {
    // Grab the topic and account to sign from the last pairing event
    const signer = this.getSigner();

    const transferTokenTransaction = await new TransferTransaction()
      .addTokenTransfer(tokenId, signer.getAccountId(), -amount)
      .addTokenTransfer(tokenId, toAddress, amount)
      .freezeWithSigner(signer);

    const txResult = await transferTokenTransaction.executeWithSigner(signer);
    return txResult.transactionId;
  }

  async transferNonFungibleToken(toAddress: AccountId, tokenId: TokenId, serialNumber: number) {
    // Grab the topic and account to sign from the last pairing event
    const signer = this.getSigner();

    const transferTokenTransaction = await new TransferTransaction()
      .addNftTransfer(tokenId, serialNumber, signer.getAccountId(), toAddress)
      .freezeWithSigner(signer);

    const txResult = await transferTokenTransaction.executeWithSigner(signer);
    return txResult.transactionId;
  }

  async associateToken(tokenId: TokenId) {
    const signer = this.getSigner();

    const associateTokenTransaction = await new TokenAssociateTransaction()
      .setAccountId(signer.getAccountId())
      .setTokenIds([tokenId])
      .freezeWithSigner(signer);

    const txResult = await associateTokenTransaction.executeWithSigner(signer);
    return txResult.transactionId;
  }

  // Purpose: build contract execute transaction and send to hashconnect for signing and execution
  // Returns: Promise<TransactionId | null>
  async executeContractFunction(contractId: ContractId, functionName: string, functionParameters: ContractFunctionParameterBuilder, gasLimit: number, amount?: Hbar) {
    const signer = this.getSigner();
    
    const tx = new ContractExecuteTransaction()
      .setContractId(contractId)
      .setGas(gasLimit)
      .setFunction(functionName, functionParameters.buildHAPIParams());

    // If amount is provided, set it as the payment amount
    if (amount) {
      tx.setPayableAmount(amount);
    }
    
    const txFrozen = await tx.freezeWithSigner(signer);
    const txResponse = await txFrozen.executeWithSigner(signer);

    // in order to read the contract call results, you will need to query the contract call's results form a mirror node using the transaction id
    // after getting the contract call results, use ethers and abi.decode to decode the call_result
    return txFrozen.transactionId;
  }

  // Purpose: takes a prepared transaction in hex encoding and signs it with the wallet
  // Returns: Promise<TransactionId | null>
  async executeTransaction(transactionBytes: string) {
    const signer = this.getSigner();

    // decode the transaction bytes
    const bytes = Buffer.from(transactionBytes, "hex");

    // create a transaction from the bytes and execute it with the signer
    const transaction = Transaction.fromBytes(bytes);
    const txResult = await transaction.executeWithSigner(signer);
    
    return txResult.transactionId;

  }

  // disconnect wallet
  disconnect() {
    const pairingData = hashConnect.hcData.pairingData[hashConnect.hcData.pairingData.length - 1];
    hashConnect.disconnect(pairingData.topic);
  }

};

export const hashConnectWallet = new HashConnectWallet();

const getPairingInfo = () => {
  if (hashConnect.hcData.pairingData.length > 0) {
    return hashConnect.hcData.pairingData[hashConnect.hcData.pairingData.length - 1];
  }
}

// set the necessary metadata for your app
// call hashconnects init function which will return your pairing code & any previously connected pariaings
// this will also start the pairing event listener
const hashConnectInitPromise = new Promise(async (resolve) => {
  /* this metadata is used to display the app so the 
      wallet can display what app is requesting access from the user
  */
  const appMetadata: HashConnectTypes.AppMetadata = {
    name: "Renderhive Service App",
    description: "The first blockchain-based crowdrender network for Blender",
    icon: new URL(RenderhiveLogo, import.meta.url).href // window.location.origin + "/logo192.png"
  };
  const initResult = await hashConnect.init(appMetadata, hederaNetwork, true)
  resolve(initResult);
});

// this component will sync the hashconnect state with the context
export const HashConnectClient = () => {
  const { setLoading } = useLoading();
  // use the HashpackContext to keep track of the hashpack account and connection
  const { setAccountId, setIsConnected } = useContext(HashconnectContext);

  // sync the hashconnect state with the context
  const syncWithHashConnect = useCallback(() => {
    const accountId = getPairingInfo()?.accountIds[0];

    if (accountId) {
      setAccountId(accountId);
      setIsConnected(true);
    } else {
      setAccountId('');
      setIsConnected(false);
    }
    
  }, [setAccountId, setIsConnected]);

  useEffect(() => {
    
    // set the loader status
    setLoading(true)

    // when the component renders, sync the hashconnect state with the context
    syncWithHashConnect();
    
    // when hashconnect is initialized, sync the hashconnect state with the context
    hashConnectInitPromise.then(() => {
      syncWithHashConnect();

      // reset the loader status
      setLoading(false)

    });

    // when pairing an account, sync the hashconnect state with the context
    hashConnect.pairingEvent.on(syncWithHashConnect);

    // when the connection status changes, sync the hashconnect state with the context
    hashConnect.connectionStatusChangeEvent.on(syncWithHashConnect)

    return () => {
      // remove the event listeners when the component unmounts
      hashConnect.pairingEvent.off(syncWithHashConnect);
      hashConnect.connectionStatusChangeEvent.off(syncWithHashConnect);

      // reset the loader status
      setLoading(false)
    }
  }, [syncWithHashConnect]);
  return null;
};
