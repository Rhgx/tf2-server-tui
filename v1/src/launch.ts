import open from 'open';

export async function launchTF2(ip: string, port: number): Promise<void> {
    const connectUrl = `steam://connect/${ip}:${port}`;
    await open(connectUrl);
}
