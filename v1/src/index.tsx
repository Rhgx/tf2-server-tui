#!/usr/bin/env node
import React from 'react';
import { render } from 'ink';
import { App } from './App.js';

const args = process.argv.slice(2);
const configPath = args.find((arg) => !arg.startsWith('-'));

// Clear the terminal
process.stdout.write('\x1Bc');

console.log('Starting TF2 Server Browser...\n');

render(<App configPath={configPath} />);
