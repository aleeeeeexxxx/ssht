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

      <Container maxWidth="lg" sx={{ mt: 3 }}>
        <Box sx={{ mb: 3 }}>
          <TunnelList
            tunnels={tunnels}
            onRefresh={fetchTunnels}
            onEdit={handleEdit}
          />
        </Box>
        <LogViewer />
      </Container>

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
