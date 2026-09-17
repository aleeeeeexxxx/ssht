import { useState, useEffect, useCallback } from 'react';
import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  CssBaseline,
  ThemeProvider,
  createTheme,
} from '@mui/material';
import { Add as AddIcon, Refresh as RefreshIcon } from '@mui/icons-material';
import TunnelList from './components/TunnelList';
import TunnelForm from './components/TunnelForm';
import LogViewer from './components/LogViewer';
import type { Tunnel } from './types/tunnel';
import * as api from './api/tunnels';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
  },
  typography: {
    fontSize: 14,
    body1: {
      '@media (max-width:600px)': {
        fontSize: '0.875rem',
      },
    },
    body2: {
      '@media (max-width:600px)': {
        fontSize: '0.75rem',
      },
    },
  },
  components: {
    MuiTableCell: {
      styleOverrides: {
        root: {
          '@media (max-width:600px)': {
            padding: '8px 4px',
            fontSize: '0.75rem',
          },
        },
      },
    },
  },
});

function App() {
  const [tunnels, setTunnels] = useState<Tunnel[]>([]);
  const [formOpen, setFormOpen] = useState(false);
  const [editTunnel, setEditTunnel] = useState<Tunnel | null>(null);

  const fetchTunnels = useCallback(async () => {
    try {
      const data = await api.listTunnels();
      setTunnels(data);
    } catch (err) {
      console.error('Failed to fetch tunnels:', err);
    }
  }, []);

  useEffect(() => {
    fetchTunnels();
    const interval = setInterval(fetchTunnels, 3000);
    return () => clearInterval(interval);
  }, [fetchTunnels]);

  const handleAdd = () => {
    setEditTunnel(null);
    setFormOpen(true);
  };

  const handleEdit = (tunnel: Tunnel) => {
    setEditTunnel(tunnel);
    setFormOpen(true);
  };

  const handleFormClose = () => {
    setFormOpen(false);
    setEditTunnel(null);
  };

  const handleFormSuccess = () => {
    fetchTunnels();
  };

  return (
    <ThemeProvider theme={darkTheme}>
      <CssBaseline />
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            ssht - SSH Tunnel Manager
          </Typography>
          <Button
            color="inherit"
            startIcon={<RefreshIcon />}
            onClick={fetchTunnels}
            sx={{ mr: 1 }}
          >
            Refresh
          </Button>
          <Button
            color="inherit"
            variant="outlined"
            startIcon={<AddIcon />}
            onClick={handleAdd}
          >
            Add Tunnel
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth={false} sx={{ mt: 2, px: { xs: 2, md: '10%' } }}>
        <Box
          sx={{
            display: 'flex',
            flexDirection: { xs: 'column', md: 'row' },
            gap: 2,
            height: { md: 'calc(100vh - 140px)' },
          }}
        >
          <Box sx={{ flex: { xs: '1', md: '1 1 60%' }, minWidth: 0, height: { xs: 'auto', md: '100%' }, overflow: 'auto' }}>
            <TunnelList
              tunnels={tunnels}
              onRefresh={fetchTunnels}
              onEdit={handleEdit}
            />
          </Box>
          <Box sx={{ flex: { xs: '1', md: '1 1 40%' }, minWidth: 0, height: { xs: '50vh', md: '100%' } }}>
            <LogViewer />
          </Box>
        </Box>
      </Container>

      <Box
        component="footer"
        sx={{
          py: 1,
          textAlign: 'center',
          color: 'grey.600',
          fontSize: '0.75rem',
        }}
      >
        © 2026 ssht. Made with ❤️
      </Box>

      <TunnelForm
        open={formOpen}
        tunnel={editTunnel}
        onClose={handleFormClose}
        onSuccess={handleFormSuccess}
      />
    </ThemeProvider>
  );
}

export default App;
