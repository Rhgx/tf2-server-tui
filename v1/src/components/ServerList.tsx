import React from 'react';
import { Box, Text } from 'ink';
import type { ServerInfo } from '../query.js';

interface ServerListProps {
    servers: ServerInfo[];
    selectedIndex: number;
}

function getPingColor(ping: number): string {
    if (ping < 50) return 'green';
    if (ping < 100) return 'yellow';
    if (ping < 150) return 'magenta';
    return 'red';
}

function formatPlayers(players: number, maxPlayers: number): string {
    return `${players.toString().padStart(2)}/${maxPlayers.toString().padEnd(2)}`;
}

function truncate(str: string, maxLength: number): string {
    if (str.length <= maxLength) return str.padEnd(maxLength);
    return str.slice(0, maxLength - 3) + '...';
}

export function ServerList({ servers, selectedIndex }: ServerListProps) {
    const nameWidth = 40;
    const playersWidth = 7;
    const pingWidth = 6;
    const mapWidth = 20;

    return (
        <Box flexDirection="column" borderStyle="single" borderColor="gray">
            {/* Header Row */}
            <Box paddingX={1} borderBottom>
                <Text bold color="white">
                    {'  '}
                    {truncate('Server Name', nameWidth)}
                    {'  '}
                    {'Players'.padEnd(playersWidth)}
                    {'  '}
                    {'Ping'.padEnd(pingWidth)}
                    {'  '}
                    {'Map'.padEnd(mapWidth)}
                </Text>
            </Box>
            <Box paddingX={1}>
                <Text dimColor>
                    {'  '}
                    {'-'.repeat(nameWidth)}
                    {'  '}
                    {'-'.repeat(playersWidth)}
                    {'  '}
                    {'-'.repeat(pingWidth)}
                    {'  '}
                    {'-'.repeat(mapWidth)}
                </Text>
            </Box>

            {/* Server Rows */}
            {servers.length === 0 ? (
                <Box paddingX={1} paddingY={1} justifyContent="center">
                    <Text dimColor>No servers configured. Edit servers.json to add servers.</Text>
                </Box>
            ) : (
                servers.map((server, index) => {
                    const isSelected = index === selectedIndex;
                    const displayName = server.label || server.name;

                    if (!server.online) {
                        return (
                            <Box
                                key={`${server.ip}:${server.port}`}
                                paddingX={1}
                                backgroundColor={isSelected ? 'gray' : undefined}
                            >
                                <Text color={isSelected ? 'white' : 'gray'}>
                                    {isSelected ? '>' : ' '}{' '}
                                    {truncate(displayName, nameWidth)}
                                    {'  '}
                                    <Text color="red">{'OFFLINE'.padEnd(playersWidth + pingWidth + mapWidth + 6)}</Text>
                                </Text>
                            </Box>
                        );
                    }

                    return (
                        <Box
                            key={`${server.ip}:${server.port}`}
                            paddingX={1}
                            backgroundColor={isSelected ? 'blueBright' : undefined}
                        >
                            <Text color={isSelected ? 'white' : undefined} bold={isSelected}>
                                {isSelected ? '>' : ' '}{' '}
                                {truncate(displayName, nameWidth)}
                                {'  '}
                                <Text color={server.players > 0 ? 'green' : 'gray'}>
                                    {formatPlayers(server.players, server.maxPlayers)}
                                </Text>
                                {'  '}
                                <Text color={getPingColor(server.ping)}>
                                    {`${server.ping}ms`.padEnd(pingWidth)}
                                </Text>
                                {'  '}
                                <Text color="cyan">{truncate(server.map, mapWidth)}</Text>
                            </Text>
                        </Box>
                    );
                })
            )}
        </Box>
    );
}
