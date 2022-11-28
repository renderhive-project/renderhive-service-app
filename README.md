# Renderhive Service App

## About the Renderhive Project

This project was created to establish the first fully decentralized crowdrendering platform for [Blender](https://www.blender.org) built on the Web3 technologies of [Hedera Hashgraph](https://www.hedera.com/) and [IPFS](https://ipfs.tech) / [Filecoin](https://filecoin.io). The aim is to gather the latent CPU/GPU power of the Blender community by creating a free marketplace where Blender artists can buy and sell rendering time on each others computers in times they don't need it for their own projects.

Please visit [https://www.renderhive.io](https://www.renderhive.io) to learn more about the project.

### Contact

* **Project Coordination:** Christian Stolze
* **Contact:** contact@renderhive.io

## About this repository

This repository contains the open source code of the Renderhive Service App. This is the central app that will run on each node of the Renderhive network.

### Contributing

If you want to contribute, feel free to create a pull request. Please note that the app currently runs on the Hedera testnet and, thus, expects a `testnet.env` file in the `hedera` folder. In order to test the app, [create a Hedera testnet account](https://docs.hedera.com/guides/testnet/testnet-access) for yourself and add the `testnet.env` to your local copy of this repository. The file content should contain the following:

```
TESTNET_ACCOUNT_ID=$YOUR_ID$
TESTNET_PRIVATE_KEY=$YOUR_PRIVATE_KEY$
```

When creating pushing commits, make sure that this file is not included, since it contains your private account details.
