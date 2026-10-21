import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GitHubClient } from '../../src/scm/github.js';

describe('GitHubClient', () => {
  let client: GitHubClient;

  beforeEach(() => {
    client = new GitHubClient({
      owner: 'owner',
      repo: 'repo',
      token: 'test-token',
      baseURL: 'https://api.github.com',
    });
  });

  describe('getFileContent', () => {
    it('encodes path segments to prevent path traversal', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
        new Response('file content', { status: 200 }),
      );

      await client.getFileContent('%2e%2e/%2e%2e/secret', '');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.github.com/repos/owner/repo/contents/%252e%252e/%252e%252e/secret',
      );

      fetchSpy.mockRestore();
    });

    it('encodes the ref query parameter', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
        new Response('file content', { status: 200 }),
      );

      await client.getFileContent('src/file.ts', 'feature/branch&inject=true');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.github.com/repos/owner/repo/contents/src/file.ts?ref=feature%2Fbranch%26inject%3Dtrue',
      );

      fetchSpy.mockRestore();
    });

    it('preserves normal file paths', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
        new Response('file content', { status: 200 }),
      );

      await client.getFileContent('src/main/app.ts', 'main');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.github.com/repos/owner/repo/contents/src/main/app.ts?ref=main',
      );

      fetchSpy.mockRestore();
    });

    it('omits ref query param when empty', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
        new Response('file content', { status: 200 }),
      );

      await client.getFileContent('file.ts', '');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.github.com/repos/owner/repo/contents/file.ts',
      );
      expect(calledUrl).not.toContain('?ref=');

      fetchSpy.mockRestore();
    });
  });
});
