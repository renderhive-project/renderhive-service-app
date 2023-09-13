import AppRouter from './AppRouter';
import { LoadingContextProvider } from './contexts/LoaderContext';
import { AllWalletsProvider } from './services/wallets/AllWalletsProvider';

// styles & themes
import "./styles/global.scss"
import { ColorModeContext, useMode } from './theme';
import { ThemeProvider } from '@emotion/react';

function App() {
  const [theme, colorMode] = useMode();

  return (
    <LoadingContextProvider>
      <ColorModeContext.Provider value={colorMode}>
        <ThemeProvider theme={theme}>
          <AllWalletsProvider>
            <AppRouter />
          </AllWalletsProvider>
        </ThemeProvider>
      </ColorModeContext.Provider>
    </LoadingContextProvider>
  )
}

export default App
