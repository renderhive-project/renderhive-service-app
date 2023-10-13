import { createBrowserRouter, RouterProvider, Outlet, Navigate, useLocation } from "react-router-dom";
import { Box, CircularProgress, CssBaseline, Toolbar, Typography } from "@mui/material";
import { useWalletInterface } from "./services/wallets/useWalletInterface";
import { useLoading } from "./contexts/LoaderContext";

// components
import Navbar from "./components/navbar/Navbar";
// import Footer from "./components/footer/Footer";
import Sidebar from "./components/sidebar/Sidebar";

// pages
import SignUp from "./pages/signup/signup";
import SignIn from "./pages/signin/signin";
import Dashboard from "./pages/dashboard/Dashboard";
import Users from "./pages/users/Users";
import Products from "./pages/products/Products";
import { useContext } from "react";
import { SessionContext, useSession } from "./contexts/SessionContext";

export default function AppRouter() {
  const { signedIn, operatorInfo, nodeInfo } = useSession();
  const { accountId } = useWalletInterface();
  const { isLoading, setLoading } = useLoading();
  // console.log("Signed in:", signedIn)
  // console.log(operatorInfo, nodeInfo, accountId)

  const Layout = () => {
    const location = useLocation();
    // console.log(location.pathname)

    return (
      <Box sx={{ display: 'flex' }}>
        <CssBaseline />

        {/* Render the app bar */}
        <Navbar />

        {/* Render the sidebar (if not on sign/signup pages) */}
        {(!["/signin", '/signup'].includes(location.pathname)) && <Sidebar />}

        {/* Render content */}
        <Box component="main" sx={{ flexGrow: 1, p: 3 }}>

          <Toolbar />

          <Box className="contentContainer">
            {/* <QueryClientProvider client={queryClient}> */}
            { ( isLoading ) ? (
              <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="80vh">  
                <CircularProgress />
              </Box>
            ) : (
              <Box height="80vh"> 
                <Outlet />
              </Box>
            )}
            {/* </QueryClientProvider> */}
          </Box>
        </Box>
      </Box>

    );
  };

  const router = createBrowserRouter([
    {
      path: "/",
      element: <Layout />,
      children: [
        {
          path: "/",
          element: ((accountId && signedIn) ? <Dashboard /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/users",
          element: ((accountId && signedIn) ? <Users /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/products",
          element: ((accountId && signedIn) ? <Products /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/signup",
          element: ((accountId && signedIn) ? <Navigate to="/" replace /> : (
            
            ((operatorInfo && operatorInfo.accountId == accountId) && (nodeInfo && nodeInfo.accountId != "")) ? <Navigate to="/signin" replace /> : <SignUp /> 
            
          )),
        },
        {
          path: "/signin",
          element: ((accountId && signedIn) ? <Navigate to="/" replace /> : (

            ((operatorInfo && operatorInfo.accountId == '0.0.0' || (operatorInfo && operatorInfo.accountId == accountId) && (nodeInfo && nodeInfo.accountId == ""))) ? <Navigate to="/signup" replace /> : <SignIn /> 
            
          )),
        },
        {
          path: "*",
          element: ((accountId && signedIn) ? <Navigate to="/" replace /> : <Navigate to="/signin" replace />),
        },
        // {
        //   path: "/users/:id",
        //   element: <User />,
        // },
        // {
        //   path: "/products/:id",
        //   element: <Product />,
        // },
      ],
    },
  ]);

  return (
    <RouterProvider router={router} />
  )
}