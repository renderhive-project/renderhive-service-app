import { PaletteMode, createTheme } from "@mui/material";
import { createContext, useEffect, useMemo, useState } from "react";

// color design tokens
export const tokens = (mode: PaletteMode) => ({
  ...(mode === 'light'

  // light mode colors
  ? {
      white: {
        100: "#323232",
        200: "#656565",
        300: "#979797",
        400: "#cacaca",
        500: "#fcfcfc",
        600: "#fdfdfd",
        700: "#fdfdfd",
        800: "#fefefe",
        900: "#fefefe",
      },
      grey: {
        100: "#141414",
        200: "#292929",
        300: "#3d3d3d",
        400: "#525252",
        500: "#666666",
        600: "#858585",
        700: "#a3a3a3",
        800: "#c2c2c2",
        900: "#e0e0e0",
      },
      primary: {
        100: "#322607",
        200: "#644c0e",
        300: "#977316",
        400: "#c9991d",
        500: "#fbbf24",
        600: "#fccc50",
        700: "#fdd97c",
        800: "#fde5a7",
        900: "#fefadb",
      },
      secondary: {
        100: "#312002",
        200: "#623f05",
        300: "#935f07",
        400: "#c47e0a",
        500: "#f59e0c",
        600: "#f7b13d",
        700: "#f9c56d",
        800: "#fbd89e",
        900: "#fdecce",
      },

    // dark mode colors
    } : {
      white: {
        100: "#fefefe",
        200: "#fefefe",
        300: "#fdfdfd",
        400: "#fdfdfd",
        500: "#fcfcfc",
        600: "#cacaca",
        700: "#979797",
        800: "#656565",
        900: "#323232",
      },
      grey: {
        100: "#e0e0e0",
        200: "#c2c2c2",
        300: "#a3a3a3",
        400: "#858585",
        500: "#666666",
        600: "#525252",
        700: "#3d3d3d",
        800: "#181818",
        900: "#141414",
      },
      primary: {
        100: "#fef2d3",
        200: "#fde5a7",
        300: "#fdd97c",
        400: "#fccc50",
        500: "#fbbf24",
        600: "#c9991d",
        700: "#977316",
        800: "#644c0e",
        900: "#322607",
      },
      secondary: {
        100: "#fdecce",
        200: "#fbd89e",
        300: "#f9c56d",
        400: "#f7b13d",
        500: "#f59e0c",
        600: "#c47e0a",
        700: "#935f07",
        800: "#623f05",
        900: "#312002"
      },

    })
})



// theme settings to return color tokens dynamically based on the mode
export const createRenderhiveTheme = (mode: PaletteMode) => {
  const colors = tokens(mode);

  // light theme definition
  return (mode === 'light' ? createTheme({
    typography: {
      fontFamily: '"Styrene A Web", "Helvetica Neue", Sans-Serif',
  
      // Font for Renderhive Logo
      h6: {
        fontFamily: '"Mono", Sans-Serif'
      },
    },
    palette: {
      mode: 'light',
      primary: {
        dark: "#F6C849",
        main: "#f1a83b",
        light: "#f8d64b",
        contrastText: "#01080D",
      },
      background: {
        default: "#E5ECEF",
        paper: "#fcfcfc",
      },
      text: {
        primary: "#000000",
        secondary: "#050F15",
        disabled: "#ececec",
      }
    },
    components: {
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundColor: colors.white[500],
            color: colors.grey[100],
          }
        }
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundColor: colors.white[500],
            color: colors.grey[100],

            '& .MuiListSubheader-root': {
              backgroundColor: 'inherit',
              color: colors.grey[100],
            }
          }
        }
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            '&:hover': {
              backgroundColor: colors.grey[800], // This sets a slight black overlay on hover
            }
          }
        }
      },
    }

  // dark theme definition
  }) : createTheme({
  
      typography: {
        fontFamily: '"Styrene A Web", "Helvetica Neue", Sans-Serif',
    
        // Font for Renderhive Logo
        h6: {
          fontFamily: '"Mono", Sans-Serif'
        },
      },
      palette: {
        mode: 'dark',
        primary: {
          dark: "#f1a83b",
          main: "#F6C849",
          light: "#f8d64b",
          contrastText: "#01080D",
        },
        background: {
          default: "#01080D",
          paper: "#050F15"
        },
        text: {
          primary: "#7f878d",
          secondary: "#050F15",
          disabled: "#ececec",
        }
      },
      components: {
        MuiAppBar: {
          styleOverrides: {
            root: {
              //backgroundColor: "#050F15",
              //text: colors.white[500],
            }
          }
        },
        MuiDrawer: {
          styleOverrides: {
            paper: {
              //backgroundColor: colors.grey[800],
              //color: colors.white[500],
    
              '& .MuiListSubheader-root': {
                //backgroundColor: 'inherit',
                //color: colors.white[500],
              }
            }
          }
        },
        MuiIconButton: {
          styleOverrides: {
            root: {
              '&:hover': {
                //backgroundColor: colors.grey[800],
              }
            }
          }
        },
      }

  })

  )

}

// define the light and dark theme
export const lightTheme = createRenderhiveTheme('light');
export const darkTheme = createRenderhiveTheme('dark');

// create a context for the color mode
export const ColorModeContext = createContext<any>({ toggleColorMode: () => {} });

// create a hook for the color mode and theme
export const  useMode = () => {
   // Initialize the mode from local storage or default to 'light'
   const [mode, setMode] = useState<PaletteMode>(
    () => window.localStorage.getItem('themeMode') as PaletteMode || 'light'
  );

  useEffect(() => {
    // Whenever mode changes, update the value in local storage
    window.localStorage.setItem('themeMode', mode);
  }, [mode]);
  
  const colorMode = useMemo(
    () => ({
      toggleColorMode: () => 
        setMode((prev) => (prev === 'light' ? 'dark' : 'light'))
    }),
    []
  );

  const theme = useMemo(() => (mode === "light" ? lightTheme : darkTheme), [mode])

  return [theme, colorMode]
};


