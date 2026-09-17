export interface Tunnel {
  name: string;
  host: string;
  port: number;
  user: string;
  auth_method: 'key' | 'password';
  key_path?: string;
  password?: string;
  remote_host: string;
  remote_port: number;
  local_host: string;
  local_port: number;
  state: TunnelState;
  error?: string;
}

export type TunnelState = 'Stopped' | 'Connecting' | 'Connected' | 'Reconnecting' | 'Error';

export interface CreateTunnelRequest {
  name: string;
  host: string;
  port?: number;
  user: string;
  auth_method: 'key' | 'password';
  key_path?: string;
  key_content?: string;
  password?: string;
  remote_host?: string;
  remote_port: number;
  local_host?: string;
  local_port: number;
}

export interface TunnelsResponse {
  tunnels: Tunnel[];
}

export interface MessageResponse {
  message: string;
  error?: string;
}

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  fields?: Record<string, unknown>;
}
