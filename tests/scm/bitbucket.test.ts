import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BitbucketClient } from '../../src/scm/bitbucket.js';

describe('BitbucketClient', () => {
  let client: BitbucketClient;

  beforeEach(() => {
    client = new BitbucketClient({
      workspace: 'ws',
      repoSlug: 'repo',
      token: 'test-token',
      baseURL: 'https://api.bitbucket.org/2.0',
    });
  });

  describe('getFileContent', () => {
    it('encodes path segments to prevent path traversal', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('file content', { status: 200 }));

      await client.getFileContent('%2e%2e/%2e%2e/secret', 'HEAD');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.bitbucket.org/2.0/repositories/ws/repo/src/HEAD/%252e%252e/%252e%252e/secret',
      );

      fetchSpy.mockRestore();
    });

    it('encodes the ref parameter', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('file content', { status: 200 }));

      await client.getFileContent('src/file.ts', 'feature/branch name');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe(
        'https://api.bitbucket.org/2.0/repositories/ws/repo/src/feature%2Fbranch%20name/src/file.ts',
      );

      fetchSpy.mockRestore();
    });

    it('preserves normal file paths', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('file content', { status: 200 }));

      await client.getFileContent('src/main/app.ts', 'main');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toBe('https://api.bitbucket.org/2.0/repositories/ws/repo/src/main/src/main/app.ts');

      fetchSpy.mockRestore();
    });

    it('defaults ref to HEAD when empty', async () => {
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('file content', { status: 200 }));

      await client.getFileContent('file.ts', '');

      const calledUrl = fetchSpy.mock.calls[0][0] as string;
      expect(calledUrl).toContain('/src/HEAD/file.ts');

      fetchSpy.mockRestore();
    });
  });
});
