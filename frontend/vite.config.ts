import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import basicSsl from '@vitejs/plugin-basic-ssl'
// import { nodePolyfills } from "vite-plugin-node-polyfills";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(), 
    basicSsl(),
    // nodePolyfills({
    //   protocolImports: true,
    //   // globals: {
    //   //   global: true,
    //   // },
    // }),
  ],
  server: { 
    https: true 
  },

  optimizeDeps: {
    esbuildOptions: {
      // Node.js global to browser globalThis
      define: {
        global: "globalThis",
      },
    },
  },

  build: {
      commonjsOptions: {
        transformMixedEsModules: true,
      },
    },
})
