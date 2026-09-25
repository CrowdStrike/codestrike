import pino from 'pino';

export function createLogger(level: string): pino.Logger {
  return pino({
    level: level || 'info',
    transport: {
      target: 'pino-pretty',
      options: {
        colorize: true,
        translateTime: 'SYS:standard',
        ignore: 'pid,hostname',
      },
    },
  });
}

export type Logger = pino.Logger;
