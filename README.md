# Renderhive Service App

## About the Renderhive Project

This project was created to establish the first fully decentralized crowdrendering platform for [Blender](https://www.blender.org) built on the Web3 technologies of [Hedera Hashgraph](https://www.hedera.com/) and [IPFS](https://ipfs.tech) / [Filecoin](https://filecoin.io). The aim is to gather the latent CPU/GPU power of the Blender community by creating a free marketplace where Blender artists can buy and sell rendering time on each others computers in times they don't need it for their own projects.

Please visit [https://www.renderhive.io](https://www.renderhive.io) to learn more about the project.

### Contact

* **Project Coordination:** Christian Stolze
* **Contact:** contact@renderhive.io

## About this repository

This repository contains the open source code of the Renderhive Service App. This is the central app that will run on each node of the Renderhive network.

### Get started

#### 1. Install Docker

Renderhive is shipped as a containerized solution. This ensure that the Renderhive software runs consistently on different systems, reduces platform-dependent bugs, contains all required dependencies, and simplifies maintanence. Please download Docker from the official website:

https://www.docker.com

#### 2. Clone the repository

```bash
git clone https://github.com/renderhive-projects/renderhive-service-app.git
cd renderhive/service_app
```

#### 3. Create a Hedera testnet account

The app currently runs on the Hedera testnet only and, thus, expects a `testnet.env` file in the `backend/src/hedera` folder, which contains your account credentials. Therefore, you need to [create a Hedera testnet account](https://docs.hedera.com/guides/testnet/testnet-access) for yourself and add the `testnet.env` to your local copy of this repository. The file should contain the following two lines:

```
TESTNET_ACCOUNT_ID=$YOUR_ID$
TESTNET_PRIVATE_KEY=$YOUR_PRIVATE_KEY$
```

`$YOUR_ID$` and `$YOUR_PRIVATE_KEY$` need to be replaces with your account details.

#### 4. Create the docker images and the container

You can use the Docker Desktop app to create the containers or simply the command-line interface:

```bash
docker-compose -p renderhive build
```

#### 4. Start the renderhive container

To spin up the renderhive service app, you run the docker containers via `docker-compose`

```bash
docker-compose -p renderhive up
```

### Contributing

If you want to contribute, feel free to create a pull request. When creating pushing commits, make sure that the `testnet.env` file is not included, since it contains your private testnet account details.
