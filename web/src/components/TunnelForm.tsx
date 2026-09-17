import { useState, useEffect, useRef } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Box,
  Alert,
  Typography,
} from '@mui/material';
import { CloudUpload as UploadIcon } from '@mui/icons-material';
import type { Tunnel, CreateTunnelRequest } from '../types/tunnel';
import * as api from '../api/tunnels';

interface TunnelFormProps {
  open: boolean;
  tunnel: Tunnel | null;
  onClose: () => void;
  onSuccess: () => void;
}

const defaultForm: CreateTunnelRequest = {
  name: '',
  host: '',
  port: 22,
  user: '',
  auth_method: 'key',
  key_path: '',
  password: '',
  remote_host: '0.0.0.0',
  remote_port: 8080,
  local_host: '127.0.0.1',
  local_port: 3000,
};

export default function TunnelForm({ open, tunnel, onClose, onSuccess }: TunnelFormProps) {
  const [form, setForm] = useState<CreateTunnelRequest>(defaultForm);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [keyFileName, setKeyFileName] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (tunnel) {
      setForm({
        name: tunnel.name,
        host: tunnel.host,
        port: tunnel.port,
        user: tunnel.user,
        auth_method: tunnel.auth_method,
        key_path: tunnel.key_path || '',
        password: tunnel.password || '',
        remote_host: tunnel.remote_host,
        remote_port: tunnel.remote_port,
        local_host: tunnel.local_host,
        local_port: tunnel.local_port,
      });
      setKeyFileName('');
    } else {
      setForm(defaultForm);
      setKeyFileName('');
    }
    setError(null);
    setErrors({});
  }, [tunnel, open]);

  const handleChange = (field: keyof CreateTunnelRequest) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    setForm((prev) => ({ ...prev, [field]: e.target.value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: false }));
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target?.result as string;
      setForm((prev) => ({ ...prev, key_content: content, key_path: '' }));
      setKeyFileName(file.name);
      if (errors.key) {
        setErrors((prev) => ({ ...prev, key: false }));
      }
    };
    reader.readAsText(file);
  };

  const handleSubmit = async () => {
    // Validate required fields
    const newErrors: Record<string, boolean> = {};
    if (!form.name.trim()) newErrors.name = true;
    if (!form.host.trim()) newErrors.host = true;
    if (!form.user.trim()) newErrors.user = true;
    if (form.auth_method === 'key' && !keyFileName && !form.key_path) newErrors.key = true;
    if (form.auth_method === 'password' && !form.password) newErrors.password = true;

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setError('Please fill in all required fields');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const payload = {
        ...form,
        port: Number(form.port),
        remote_port: Number(form.remote_port),
        local_port: Number(form.local_port),
      };
      await api.createTunnel(payload);
      onSuccess();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{tunnel ? 'Edit Tunnel' : 'Add Tunnel'}</DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
          <TextField
            label="Name"
            fullWidth
            required
            value={form.name}
            onChange={handleChange('name')}
            disabled={!!tunnel}
            error={errors.name}
            helperText={errors.name ? 'Name is required' : ''}
          />
          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="SSH Host"
              fullWidth
              required
              value={form.host}
              onChange={handleChange('host')}
              error={errors.host}
              helperText={errors.host ? 'Host is required' : ''}
            />
            <TextField
              label="SSH Port"
              type="number"
              sx={{ width: 120 }}
              value={form.port}
              onChange={handleChange('port')}
            />
          </Box>
          <TextField
            label="Username"
            fullWidth
            required
            value={form.user}
            onChange={handleChange('user')}
            error={errors.user}
            helperText={errors.user ? 'Username is required' : ''}
          />
          <FormControl fullWidth>
            <InputLabel>Auth Method</InputLabel>
            <Select
              value={form.auth_method}
              label="Auth Method"
              onChange={(e) => setForm((prev) => ({ ...prev, auth_method: e.target.value as 'key' | 'password' }))}
            >
              <MenuItem value="key">SSH Key</MenuItem>
              <MenuItem value="password">Password</MenuItem>
            </Select>
          </FormControl>
          {form.auth_method === 'key' ? (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileSelect}
                style={{ display: 'none' }}
                accept=".pem,.key,*"
              />
              <Button
                variant="outlined"
                color={errors.key ? 'error' : 'primary'}
                startIcon={<UploadIcon />}
                onClick={() => fileInputRef.current?.click()}
                fullWidth
                sx={errors.key ? { borderColor: 'error.main' } : {}}
              >
                {keyFileName ? `Key: ${keyFileName}` : 'Upload SSH Key *'}
              </Button>
              {errors.key ? (
                <Typography variant="caption" color="error">
                  SSH key is required
                </Typography>
              ) : !keyFileName && (
                <Typography variant="caption" color="text.secondary">
                  Select your private key file (e.g., id_rsa)
                </Typography>
              )}
            </Box>
          ) : (
            <TextField
              label="Password"
              type="password"
              fullWidth
              required
              value={form.password}
              onChange={handleChange('password')}
              error={errors.password}
              helperText={errors.password ? 'Password is required' : ''}
            />
          )}
          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="Remote Host"
              fullWidth
              value={form.remote_host}
              onChange={handleChange('remote_host')}
            />
            <TextField
              label="Remote Port"
              type="number"
              sx={{ width: 120 }}
              value={form.remote_port}
              onChange={handleChange('remote_port')}
            />
          </Box>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <TextField
              label="Local Host"
              fullWidth
              value={form.local_host}
              onChange={handleChange('local_host')}
            />
            <TextField
              label="Local Port"
              type="number"
              sx={{ width: 120 }}
              value={form.local_port}
              onChange={handleChange('local_port')}
            />
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={loading}>
          {loading ? 'Saving...' : 'Save'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
