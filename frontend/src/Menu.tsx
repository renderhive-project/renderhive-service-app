// import DefaultIcon from '@mui/icons-material/Circle';
import DashboardIcon from '@mui/icons-material/Dashboard';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import NodesIcon from '@mui/icons-material/Devices';
import RenderJobsIcon from '@mui/icons-material/WorkHistory';
import RenderResultsIcon from '@mui/icons-material/ViewInAr';


// Sidbar style settings
export const drawerWidth = 340;

// Sidebar menu items
export const menuItems = [
    {
        group: "General",
        items: [
            { text: 'Dashboard', icon: <DashboardIcon />, link: '/' },
        ]
    },
    {
        group: "Manage",
        items: [
            { text: 'Account', icon: <AccountBalanceIcon />, link: '/users'  },
            { text: 'Nodes', icon: <NodesIcon />, link: '/products'  },
        ]
    },
    {
        group: "Rendering",
        items: [
            { text: 'Render Jobs', icon: <RenderJobsIcon />, link: '/render_jobs'  },
            { text: 'Render Results', icon: <RenderResultsIcon />, link: '/render_results'  },
        ]
    },

];