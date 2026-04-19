import { GameDig } from 'gamedig';
import type { ServerConfig } from './config.js';

export interface ServerInfo {
    ip: string;
    port: number;
    label?: string;
    name: string;
    map: string;
    players: number;
    maxPlayers: number;
    ping: number;
    online: boolean;
    error?: string;
}

export async function queryServer(server: ServerConfig): Promise<ServerInfo> {
    const baseInfo: ServerInfo = {
        ip: server.ip,
        port: server.port,
        label: server.label,
        name: server.label || `${server.ip}:${server.port}`,
        map: '',
        players: 0,
        maxPlayers: 0,
        ping: 0,
        online: false,
    };

    try {
        const result = await GameDig.query({
            type: 'teamfortress2',
            host: server.ip,
            port: server.port,
            socketTimeout: 5000,
        });

        return {
            ...baseInfo,
            name: result.name || baseInfo.name,
            map: result.map || '',
            players: result.numplayers ?? result.players?.length ?? 0,
            maxPlayers: result.maxplayers ?? 0,
            ping: result.ping ?? 0,
            online: true,
        };
    } catch (e) {
        return {
            ...baseInfo,
            online: false,
            error: e instanceof Error ? e.message : 'Unknown error',
        };
    }
}

export async function queryAllServers(
    servers: ServerConfig[]
): Promise<ServerInfo[]> {
    const results = await Promise.all(servers.map(queryServer));
    return results;
}
