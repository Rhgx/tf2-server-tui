import React, { useState, useEffect, useCallback } from 'react';
import { Box, Text, useApp, useInput } from 'ink';
import { Header } from './components/Header.js';
import { Footer } from './components/Footer.js';
import { ServerList } from './components/ServerList.js';
import { loadConfig } from './config.js';
import { queryAllServers, type ServerInfo } from './query.js';
import { launchTF2 } from './launch.js';

interface AppProps {
    configPath?: string;
}

export function App({ configPath }: AppProps) {
    const { exit } = useApp();
    const [servers, setServers] = useState<ServerInfo[]>([]);
    const [selectedIndex, setSelectedIndex] = useState(0);
    const [lastRefresh, setLastRefresh] = useState<Date | null>(null);
    const [isRefreshing, setIsRefreshing] = useState(true);
    const [config] = useState(() => loadConfig(configPath));
    const [message, setMessage] = useState<string | null>(null);

    const refresh = useCallback(async () => {
        setIsRefreshing(true);
        try {
            const results = await queryAllServers(config.servers);
            setServers(results);
            setLastRefresh(new Date());
        } catch (e) {
            setMessage(`Error refreshing: ${e}`);
        } finally {
            setIsRefreshing(false);
        }
    }, [config.servers]);

    // Initial load and auto-refresh
    useEffect(() => {
        refresh();
        const interval = setInterval(refresh, config.refreshInterval * 1000);
        return () => clearInterval(interval);
    }, [refresh, config.refreshInterval]);

    // Clear message after 3 seconds
    useEffect(() => {
        if (message) {
            const timeout = setTimeout(() => setMessage(null), 3000);
            return () => clearTimeout(timeout);
        }
    }, [message]);

    useInput((input, key) => {
        if (input === 'q' || input === 'Q') {
            exit();
            return;
        }

        if (input === 'r' || input === 'R') {
            refresh();
            return;
        }

        if (key.upArrow) {
            setSelectedIndex((prev) => (prev > 0 ? prev - 1 : servers.length - 1));
            return;
        }

        if (key.downArrow) {
            setSelectedIndex((prev) => (prev < servers.length - 1 ? prev + 1 : 0));
            return;
        }

        if (key.return && servers.length > 0) {
            const server = servers[selectedIndex];
            if (server && server.online) {
                setMessage(`Connecting to ${server.name}...`);
                launchTF2(server.ip, server.port).catch((e) => {
                    setMessage(`Failed to launch: ${e}`);
                });
            } else if (server) {
                setMessage('Cannot connect to offline server');
            }
            return;
        }
    });

    return (
        <Box flexDirection="column" padding={1}>
            <Header lastRefresh={lastRefresh} isRefreshing={isRefreshing} />

            <Box marginY={1}>
                <ServerList servers={servers} selectedIndex={selectedIndex} />
            </Box>

            {message && (
                <Box justifyContent="center" marginBottom={1}>
                    <Text color="yellow">{message}</Text>
                </Box>
            )}

            <Footer />

            <Box justifyContent="center" marginTop={1}>
                <Text dimColor>
                    Auto-refresh every {config.refreshInterval} seconds
                </Text>
            </Box>
        </Box>
    );
}
