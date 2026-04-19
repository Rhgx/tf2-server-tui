import { existsSync, readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export interface ServerConfig {
    ip: string;
    port: number;
    label?: string;
}

export interface Config {
    servers: ServerConfig[];
    refreshInterval: number;
}

const DEFAULT_CONFIG: Config = {
    servers: [],
    refreshInterval: 60,
};

export function loadConfig(configPath?: string): Config {
    const paths = [
        configPath,
        resolve(process.cwd(), 'servers.json'),
        resolve(__dirname, '..', 'servers.json'),
    ].filter(Boolean) as string[];

    for (const path of paths) {
        if (existsSync(path)) {
            try {
                const content = readFileSync(path, 'utf-8');
                const parsed = JSON.parse(content);
                return {
                    servers: parsed.servers || [],
                    refreshInterval: parsed.refreshInterval || 60,
                };
            } catch (e) {
                console.error(`Failed to parse config at ${path}:`, e);
            }
        }
    }

    return DEFAULT_CONFIG;
}
