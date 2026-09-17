import { useState, useEffect, useRef } from 'react';
import {
  Paper,
  Typography,
  Box,
  IconButton,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  RadioGroup,
  FormControlLabel,
  Radio,
} from '@mui/material';
import { Delete as ClearIcon, Settings as SettingsIcon } from '@mui/icons-material';
import type { LogEntry } from '../types/tunnel';
import { getLogsWebSocketUrl, getLogLevel, setLogLevel } from '../api/tunnels';

const levelColors: Record<string, 'success' | 'info' | 'warning' | 'error' | 'default'> = {
  DEBUG: 'default',
  INFO: 'info',
  WARN: 'warning',
  ERROR: 'error',
};

export default function LogViewer() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [filter, setFilter] = useState<string>('all');
  const [serverLevel, setServerLevel] = useState<string>('info');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [tempLevel, setTempLevel] = useState<string>('info');
  const [connected, setConnected] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    getLogLevel().then((level) => {
      setServerLevel(level);
      setTempLevel(level);
    }).catch(console.error);
  }, []);

  useEffect(() => {
    const connect = () => {
      const ws = new WebSocket(getLogsWebSocketUrl());

      ws.onopen = () => {
        setConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const log: LogEntry = JSON.parse(event.data);
          setLogs((prev) => [...prev.slice(-499), log]);
        } catch {
          // Ignore parse errors
        }
      };

      ws.onclose = () => {
        setConnected(false);
        // Reconnect after 3 seconds
        setTimeout(connect, 3000);
      };

      ws.onerror = () => {
        ws.close();
      };

      wsRef.current = ws;
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  const filteredLogs = filter === 'all'
    ? logs
    : logs.filter((log) => log.level === filter);

  const handleClear = () => {
    setLogs([]);
  };

  const handleOpenSettings = () => {
    setTempLevel(serverLevel);
    setSettingsOpen(true);
  };

  const handleSaveSettings = async () => {
    try {
      await setLogLevel(tempLevel);
      setServerLevel(tempLevel);
      setSettingsOpen(false);
    } catch (err) {
      console.error('Failed to set log level:', err);
    }
  };

  return (
    <Paper sx={{ p: 2, height: '100%', minHeight: 200, display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="h6">Logs</Typography>
          <Chip
            label={connected ? 'Connected' : 'Disconnected'}
            color={connected ? 'success' : 'error'}
            size="small"
          />
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <FormControl size="small" sx={{ minWidth: 100 }}>
            <InputLabel>Filter</InputLabel>
            <Select
              value={filter}
              label="Filter"
              onChange={(e) => setFilter(e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="DEBUG">Debug</MenuItem>
              <MenuItem value="INFO">Info</MenuItem>
              <MenuItem value="WARN">Warn</MenuItem>
              <MenuItem value="ERROR">Error</MenuItem>
            </Select>
          </FormControl>
          <IconButton size="small" onClick={handleOpenSettings} title="Log Settings">
            <SettingsIcon />
          </IconButton>
          <IconButton size="small" onClick={handleClear} title="Clear Logs">
            <ClearIcon />
          </IconButton>
        </Box>
      </Box>
      <Box
        sx={{
          flex: 1,
          overflow: 'auto',
          bgcolor: '#1e1e1e',
          borderRadius: 1,
          p: 1,
          fontFamily: 'monospace',
          fontSize: '0.85rem',
        }}
      >
        {filteredLogs.length === 0 ? (
          <Typography color="grey.500" sx={{ fontFamily: 'monospace' }}>
            No logs yet...
          </Typography>
        ) : (
          filteredLogs.map((log, index) => (
            <Box key={index} sx={{ display: 'flex', gap: 1, color: '#d4d4d4' }}>
              <Typography
                component="span"
                sx={{ color: '#6a9955', fontFamily: 'monospace', fontSize: 'inherit' }}
              >
                {log.timestamp}
              </Typography>
              <Chip
                label={log.level}
                color={levelColors[log.level] || 'default'}
                size="small"
                sx={{ height: 18, fontSize: '0.7rem' }}
              />
              <Typography
                component="span"
                sx={{ fontFamily: 'monospace', fontSize: 'inherit' }}
              >
                {log.message}
                {log.fields && Object.keys(log.fields).length > 0 && (
                  <span style={{ color: '#9cdcfe' }}>
                    {' '}
                    {JSON.stringify(log.fields)}
                  </span>
                )}
              </Typography>
            </Box>
          ))
        )}
        <div ref={logsEndRef} />
      </Box>

      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)}>
        <DialogTitle>Log Settings</DialogTitle>
        <DialogContent>
          <Typography variant="subtitle2" sx={{ mb: 1 }}>
            Server Log Level
          </Typography>
          <Typography variant="caption" color="text.secondary" sx={{ mb: 2, display: 'block' }}>
            Current: {serverLevel.toUpperCase()}
          </Typography>
          <RadioGroup value={tempLevel} onChange={(e) => setTempLevel(e.target.value)}>
            <FormControlLabel value="debug" control={<Radio />} label="Debug - All logs" />
            <FormControlLabel value="info" control={<Radio />} label="Info - Info and above" />
            <FormControlLabel value="warn" control={<Radio />} label="Warn - Warnings and errors" />
            <FormControlLabel value="error" control={<Radio />} label="Error - Errors only" />
          </RadioGroup>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSettingsOpen(false)}>Cancel</Button>
          <Button onClick={handleSaveSettings} variant="contained">Save</Button>
        </DialogActions>
      </Dialog>
    </Paper>
  );
}
