import { createBrowserRouter, RouterProvider, Outlet } from "react-router-dom";

// components
import Navbar from "./components/navbar/Navbar";
// import Footer from "./components/footer/Footer";
import Sidebar from "./components/sidebar/Sidebar";

// pages
import Dashboard from "./pages/dashboard/Dashboard";
import Users from "./pages/users/Users";
import Products from "./pages/products/Products";
import Login from "./pages/login/Login";
import { Box, CssBaseline, Toolbar } from "@mui/material";

export default function AppRouter() {
  const Layout = () => {
    return (
      <Box sx={{ display: 'flex' }}>
        <CssBaseline />

        {/* Render the app bar */}
        <Navbar />

        {/* Render the sidebar */}
        <Sidebar />

        {/* Render content */}
        <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
          <Toolbar />

          <Box className="contentContainer">
            {/* <QueryClientProvider client={queryClient}> */}
            <Outlet />
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
          element: <Dashboard />,
        },
        {
          path: "/users",
          element: <Users />,
        },
        {
          path: "/products",
          element: <Products />,
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
      path: "/login",
      element: <Login />,
    },
  ]);

  return (
    <RouterProvider router={router} />
  )
}