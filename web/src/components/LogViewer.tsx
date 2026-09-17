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
} from '@mui/material';
import { Delete as ClearIcon } from '@mui/icons-material';
import type { LogEntry } from '../types/tunnel';
import { getLogsWebSocketUrl } from '../api/tunnels';

const levelColors: Record<string, 'success' | 'info' | 'warning' | 'error' | 'default'> = {
  DEBUG: 'default',
  INFO: 'info',
  WARN: 'warning',
  ERROR: 'error',
};

export default function LogViewer() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [filter, setFilter] = useState<string>('all');
  const [connected, setConnected] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

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
            <InputLabel>Level</InputLabel>
            <Select
              value={filter}
              label="Level"
              onChange={(e) => setFilter(e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="DEBUG">Debug</MenuItem>
              <MenuItem value="INFO">Info</MenuItem>
              <MenuItem value="WARN">Warn</MenuItem>
              <MenuItem value="ERROR">Error</MenuItem>
            </Select>
          </FormControl>
          <IconButton size="small" onClick={handleClear}>
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
    </Paper>
  );
}
