export interface PullRequestFile {
  filename: string;
  status: string;
  patch: string;
  content: string;
}

export interface ReviewComment {
  file: string;
  line: number;
  body: string;
}

export interface PRComment {
  id: number;
  author: string;
  body: string;
  createdAt: Date;
  path: string;
  line: number;
  inReplyTo: number;
}

export interface SCMClient {
  getPullRequestDiff(number: number): Promise<string>;
  getPullRequestFiles(number: number): Promise<PullRequestFile[]>;
  getPullRequestDescription(number: number): Promise<{ title: string; body: string }>;
  getFileContent(path: string, ref: string): Promise<string>;
  pullRequestExists(number: number): Promise<boolean>;
  publishComment(number: number, body: string): Promise<void>;
  getPRComments(number: number): Promise<PRComment[]>;
  getPRReviewComments(number: number): Promise<PRComment[]>;
}
