import type { Tunnel, TunnelsResponse, CreateTunnelRequest, MessageResponse } from '../types/tunnel';

const API_BASE = '/api';

export async function listTunnels(): Promise<Tunnel[]> {
  const res = await fetch(`${API_BASE}/tunnels`);
  if (!res.ok) throw new Error('Failed to fetch tunnels');
  const data: TunnelsResponse = await res.json();
  return data.tunnels;
}

export async function getTunnel(name: string): Promise<Tunnel> {
  const res = await fetch(`${API_BASE}/tunnels/${encodeURIComponent(name)}`);
  if (!res.ok) throw new Error('Failed to fetch tunnel');
  return res.json();
}

export async function createTunnel(tunnel: CreateTunnelRequest): Promise<MessageResponse> {
  const res = await fetch(`${API_BASE}/tunnels`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(tunnel),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to create tunnel');
  }
  return res.json();
}

export async function deleteTunnel(name: string): Promise<MessageResponse> {
  const res = await fetch(`${API_BASE}/tunnels/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to delete tunnel');
  }
  return res.json();
}

export async function startTunnel(name: string): Promise<MessageResponse> {
  const res = await fetch(`${API_BASE}/tunnels/${encodeURIComponent(name)}/start`, {
    method: 'POST',
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to start tunnel');
  }
  return res.json();
}

export async function stopTunnel(name: string): Promise<MessageResponse> {
  const res = await fetch(`${API_BASE}/tunnels/${encodeURIComponent(name)}/stop`, {
    method: 'POST',
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to stop tunnel');
  }
  return res.json();
}

export function getLogsWebSocketUrl(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/api/ws/logs`;
}

export async function getLogLevel(): Promise<string> {
  const res = await fetch(`${API_BASE}/log-level`);
  if (!res.ok) throw new Error('Failed to get log level');
  const data = await res.json();
  return data.level;
}

export async function setLogLevel(level: string): Promise<void> {
  const res = await fetch(`${API_BASE}/log-level`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ level }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Failed to set log level');
  }
}
