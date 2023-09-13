import { createBrowserRouter, RouterProvider, Outlet, Navigate, useLocation } from "react-router-dom";
import { Box, CircularProgress, CssBaseline, Toolbar } from "@mui/material";
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
import RPCTest from "./pages/rpctest/RPCTest";

export default function AppRouter() {
  const { accountId } = useWalletInterface();
  const { isLoading, setLoading } = useLoading();

  const Layout = () => {
    const location = useLocation();
    console.log(location.pathname)

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
              <Outlet />
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
          element: (accountId ? <Dashboard /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/users",
          element: (accountId ? <Users /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/products",
          element: (accountId ? <Products /> : <Navigate to="/signin" replace />),
        },
        {
          path: "/signup",
          element: (accountId ? <Navigate to="/" replace /> : <SignUp />),
        },
        {
          path: "/signin",
          element: (accountId ? <Navigate to="/" replace /> : <SignIn />),
        },
        {
          path: "*",
          element: (accountId ? <Navigate to="/" replace /> : <Navigate to="/signin" replace />),
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
    {
      path: "/test",
      element: <RPCTest />,
    },
  ]);

  return (
    <RouterProvider router={router} />
  )
}