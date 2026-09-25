import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getEnvString, getEnvBool, getEnvFloat, getEnvInt } from '../src/env.js';

describe('env', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  describe('getEnvString', () => {
    it('returns env value when set', () => {
      process.env.TEST_KEY = 'hello';
      expect(getEnvString('TEST_KEY', 'default')).toBe('hello');
    });

    it('returns default when not set', () => {
      expect(getEnvString('NONEXISTENT', 'fallback')).toBe('fallback');
    });
  });

  describe('getEnvBool', () => {
    it('returns true for "true"', () => {
      process.env.BOOL_KEY = 'true';
      expect(getEnvBool('BOOL_KEY', false)).toBe(true);
    });

    it('returns true for "1"', () => {
      process.env.BOOL_KEY = '1';
      expect(getEnvBool('BOOL_KEY', false)).toBe(true);
    });

    it('returns false for other values', () => {
      process.env.BOOL_KEY = 'no';
      expect(getEnvBool('BOOL_KEY', true)).toBe(false);
    });

    it('returns default when not set', () => {
      expect(getEnvBool('NONEXISTENT', true)).toBe(true);
    });
  });

  describe('getEnvFloat', () => {
    it('parses float value', () => {
      process.env.FLOAT_KEY = '0.75';
      expect(getEnvFloat('FLOAT_KEY', 0.5)).toBe(0.75);
    });

    it('returns default for invalid float', () => {
      process.env.FLOAT_KEY = 'abc';
      expect(getEnvFloat('FLOAT_KEY', 0.5)).toBe(0.5);
    });

    it('returns default when not set', () => {
      expect(getEnvFloat('NONEXISTENT', 1.0)).toBe(1.0);
    });
  });

  describe('getEnvInt', () => {
    it('parses int value', () => {
      process.env.INT_KEY = '42';
      expect(getEnvInt('INT_KEY', 10)).toBe(42);
    });

    it('returns default for invalid int', () => {
      process.env.INT_KEY = 'xyz';
      expect(getEnvInt('INT_KEY', 10)).toBe(10);
    });

    it('returns default when not set', () => {
      expect(getEnvInt('NONEXISTENT', 99)).toBe(99);
    });
  });
});
