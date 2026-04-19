import React from 'react';
import { Box, Text } from 'ink';

export function Footer() {
    return (
        <Box
            borderStyle="single"
            borderColor="gray"
            paddingX={1}
            justifyContent="center"
            gap={2}
        >
            <Text>
                <Text color="yellow">[Up/Down]</Text>
                <Text dimColor> Navigate</Text>
            </Text>
            <Text>
                <Text color="green">[Enter]</Text>
                <Text dimColor> Connect</Text>
            </Text>
            <Text>
                <Text color="blue">[R]</Text>
                <Text dimColor> Refresh</Text>
            </Text>
            <Text>
                <Text color="red">[Q]</Text>
                <Text dimColor> Quit</Text>
            </Text>
        </Box>
    );
}
