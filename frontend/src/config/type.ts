// Define types for the Hedera networks
// #################################################################
export type NetworkNames = "testnet" | "mainnet";
export type ChainId = '0x128';
export type NetworkConfig = {
  network: NetworkNames,
  jsonRpcUrl: string,
  mirrorNodeUrl: string,
  chainId: ChainId,
}

// purpose of this file is to define the type of the config object
export type NetworkConfigs = {
  [key in NetworkNames]: {
    network: NetworkNames,
    jsonRpcUrl: string,
    mirrorNodeUrl: string,
    chainId: ChainId,
  }
};

// Define api endpoints
// #################################################################
export type EndpointNames = "jsonrpc";

// define type
export type ApiConfigs = {
  [key in EndpointNames]: {
    endpoint: EndpointNames,
    url: string,
  }
};

// Provide an app configuration type 
// #################################################################
export type AppConfig = {
  networks: NetworkConfigs,
  api: ApiConfigs,
}
