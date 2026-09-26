export const PROVIDER_GITHUB = 'github';
export const PROVIDER_BITBUCKET = 'bitbucket';

export interface PRReference {
  provider: string;
  owner: string;
  repo: string;
  number: number;
}

export function parsePrReference(owner: string, repo: string, number: number): PRReference {
  return { provider: PROVIDER_GITHUB, owner, repo, number };
}

export function parsePrUrl(url: string): PRReference {
  url = url.replace(/\/+$/, '');
  const parts = url.split('/');

  const pullRequestsIdx = findSegment(parts, 'pull-requests');
  if (pullRequestsIdx !== -1) {
    return parsePrParts(parts, pullRequestsIdx, PROVIDER_BITBUCKET);
  }

  const pullIdx = findSegment(parts, 'pull');
  if (pullIdx !== -1) {
    return parsePrParts(parts, pullIdx, PROVIDER_GITHUB);
  }

  throw new Error('invalid PR URL: missing /pull/ or /pull-requests/ segment');
}

function parsePrParts(parts: string[], pullIdx: number, provider: string): PRReference {
  if (pullIdx < 2 || pullIdx + 1 >= parts.length) {
    throw new Error('invalid PR URL: not enough segments');
  }

  const owner = parts[pullIdx - 2];
  const repo = parts[pullIdx - 1];
  if (!owner || !repo) {
    throw new Error('invalid PR URL: missing owner or repo');
  }

  const numberStr = parts[pullIdx + 1];
  const number = parseInt(numberStr, 10);
  if (isNaN(number)) {
    throw new Error(`invalid PR number "${numberStr}"`);
  }

  return { provider, owner, repo, number };
}

function findSegment(parts: string[], segment: string): number {
  return parts.indexOf(segment);
}
