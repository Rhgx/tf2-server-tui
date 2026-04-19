import React from 'react';
import { Box, Text } from 'ink';

interface HeaderProps {
    lastRefresh: Date | null;
    isRefreshing: boolean;
}

export function Header({ lastRefresh, isRefreshing }: HeaderProps) {
    const formatTime = (date: Date) => {
        return date.toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false,
        });
    };

    return (
        <Box
            flexDirection="column"
            borderStyle="double"
            borderColor="cyan"
            paddingX={1}
        >
            <Box justifyContent="center">
                <Text bold color="cyan">
                    TF2 SERVER BROWSER
                </Text>
            </Box>
            <Box justifyContent="center">
                <Text dimColor>
                    {isRefreshing
                        ? 'Refreshing...'
                        : lastRefresh
                            ? `Last updated: ${formatTime(lastRefresh)}`
                            : 'Loading...'}
                </Text>
            </Box>
        </Box>
    );
}
