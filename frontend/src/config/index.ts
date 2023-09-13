import { networkConfig } from "./networks";
import { apiConfig } from "./api";
import { AppConfig } from "./type";
import * as constants from "./constants";

export * from "./type";

export const appConfig: AppConfig & {
  constants: typeof constants
} = {
  networks: networkConfig,
  api: apiConfig,
  constants
}
