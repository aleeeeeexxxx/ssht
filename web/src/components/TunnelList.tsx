import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Chip,
  Menu,
  MenuItem,
  Tooltip,
  CircularProgress,
} from '@mui/material';
import {
  MoreVert as MoreVertIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material';
import type { Tunnel, TunnelState } from '../types/tunnel';
import * as api from '../api/tunnels';

interface TunnelListProps {
  tunnels: Tunnel[];
  onRefresh: () => void;
  onEdit: (tunnel: Tunnel) => void;
}

const stateColors: Record<TunnelState, 'success' | 'default' | 'warning' | 'error'> = {
  Connected: 'success',
  Stopped: 'default',
  Connecting: 'warning',
  Reconnecting: 'warning',
  Error: 'error',
};

export default function TunnelList({ tunnels, onRefresh, onEdit }: TunnelListProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedTunnel, setSelectedTunnel] = useState<Tunnel | null>(null);
  const [loading, setLoading] = useState<string | null>(null);

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, tunnel: Tunnel) => {
    setAnchorEl(event.currentTarget);
    setSelectedTunnel(tunnel);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedTunnel(null);
  };

  const handleStart = async (name: string) => {
    setLoading(name);
    try {
      await api.startTunnel(name);
      onRefresh();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(null);
      handleMenuClose();
    }
  };

  const handleStop = async (name: string) => {
    setLoading(name);
    try {
      await api.stopTunnel(name);
      onRefresh();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(null);
      handleMenuClose();
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete tunnel "${name}"?`)) return;
    setLoading(name);
    try {
      await api.deleteTunnel(name);
      onRefresh();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(null);
      handleMenuClose();
    }
  };

  const handleEdit = (tunnel: Tunnel) => {
    onEdit(tunnel);
    handleMenuClose();
  };

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>SSH Server</TableCell>
            <TableCell>Remote</TableCell>
            <TableCell>Local</TableCell>
            <TableCell>Status</TableCell>
            <TableCell align="right">Actions</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {tunnels.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} align="center">
                No tunnels configured
              </TableCell>
            </TableRow>
          ) : (
            tunnels.map((tunnel) => (
              <TableRow key={tunnel.name}>
                <TableCell>{tunnel.name}</TableCell>
                <TableCell>{tunnel.user}@{tunnel.host}:{tunnel.port}</TableCell>
                <TableCell>{tunnel.remote_host}:{tunnel.remote_port}</TableCell>
                <TableCell>{tunnel.local_host}:{tunnel.local_port}</TableCell>
                <TableCell>
                  <Tooltip title={tunnel.error || ''}>
                    <Chip
                      label={tunnel.state}
                      color={stateColors[tunnel.state]}
                      size="small"
                    />
                  </Tooltip>
                </TableCell>
                <TableCell align="right">
                  {loading === tunnel.name ? (
                    <CircularProgress size={24} />
                  ) : (
                    <>
                      {tunnel.state === 'Stopped' ? (
                        <IconButton size="small" onClick={() => handleStart(tunnel.name)}>
                          <PlayIcon />
                        </IconButton>
                      ) : (
                        <IconButton size="small" onClick={() => handleStop(tunnel.name)}>
                          <StopIcon />
                        </IconButton>
                      )}
                      <IconButton size="small" onClick={(e) => handleMenuOpen(e, tunnel)}>
                        <MoreVertIcon />
                      </IconButton>
                    </>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        <MenuItem onClick={() => selectedTunnel && handleEdit(selectedTunnel)}>
          <EditIcon fontSize="small" sx={{ mr: 1 }} /> Edit
        </MenuItem>
        <MenuItem onClick={() => selectedTunnel && handleDelete(selectedTunnel.name)}>
          <DeleteIcon fontSize="small" sx={{ mr: 1 }} /> Delete
        </MenuItem>
      </Menu>
    </TableContainer>
  );
}
