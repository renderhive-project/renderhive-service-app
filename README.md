# Renderhive Service App

## About the Renderhive Project

This project was created to establish the first fully decentralized crowdrendering platform for [Blender](https://www.blender.org) built on the Web3 technologies of [Hedera Hashgraph](https://www.hedera.com/) and [IPFS](https://ipfs.tech) / [Filecoin](https://filecoin.io). The aim is to gather the latent CPU/GPU power of the Blender community by creating a free marketplace where Blender artists can buy and sell rendering time on each others computers in times they don't need it for their own projects.

Please visit [https://www.renderhive.io](https://www.renderhive.io) to learn more about the project.

### Contact

* **Project Coordination:** Christian Stolze
* **Contact:** contact@renderhive.io

## About this repository

This repository contains the open source code of the Renderhive Service App. This is the central app that will run on each node of the Renderhive network.

### Build and run the App

#### 1. Install software dependencies

Please download and install the following software, if you want to test the Renderhive software:

- [Docker](https://www.docker.com/get-started/)
- [Chrome browser](https://www.google.com/chrome/)
- [Hashpack wallet app](https://www.hashpack.app/download)

#### 2. Create a Hedera testnet account

The app currently runs on the Hedera testnet only and requires you to create a testnet account. This can be achieved [via the Hedera developer portal](https://docs.hedera.com/guides/testnet/testnet-access) or using the [Hashpack wallet](https://www.hashpack.app). 

#### 3. Clone the repository

```bash
git clone https://github.com/renderhive-projects/renderhive-service-app.git
```

#### 4. Build the docker images

Once the `docker-compose.yml` was modified, `cd` into main folder of the code to build the renderhive images using the `docker-compose` command:

```bash
cd ..
docker-compose -p renderhive build
```

Depending on your hardware, the compilation process may take a while (e.g., 5 - 10 min on an (Intel-based) MacBook Pro 2020).

#### 5. Run the docker container

To spin up the renderhive service app, you now need to start the docker containers via `docker-compose`

```bash
docker-compose -p renderhive up
```

#### 6. Renderhive webapp

Since the code base is still under heavy construction, the existing development frontend is mainly there to test important functionalities during the development. Therefore, the workflows are not streamlined yet. For example, the Renderhive app uses a self-signed SSL certificate to establish a HTTPS connection between the backend and frontend. When you open the frontend via

```bash
https://localhost:5173/
```

the browser will warn you about the self-signed (not trusted) certificate. This is just a quick-and-dirty implementation of HTTPS for development purposes and will change in later versions. So, please skip the warning for now. You should now see the development frontend.

<img width="50%" alt="renderhive_frontend_preview" src="https://github.com/renderhive-projects/renderhive-service-app/assets/3891338/1a171aaf-06d9-4644-b7cc-72ef7f0e9988">

### Contributing

If you want to contribute, feel free to create a pull request. When pushing commits, make sure that local files (e.g., your configuration files) are not included, since it contains your private testnet account details.
