export const Version = process.env.CODESTRIKE_VERSION || 'dev';
export const Commit = process.env.CODESTRIKE_COMMIT || 'unknown';
export const CommitDate = process.env.CODESTRIKE_COMMIT_DATE || 'unknown';
export const BuiltBy = process.env.CODESTRIKE_BUILT_BY || 'unknown';
export const BuildDate = process.env.CODESTRIKE_BUILD_DATE || 'unknown';

export function shortVersion(): string {
  const commit = Commit.length > 7 ? Commit.slice(0, 7) : Commit;
  return `${Version} (${commit} ${CommitDate})`;
}
