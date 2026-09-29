export type URLItem = {
  originalURL: string;
  shortURL: string;
  expiresAt?: string | null;
  clickCount?: number;
  createdAt?: string;
};

export type ShortenURLResponse = {
  shortURL: string;
  expiresAt?: string | null;
};

export type ShortenURLAPIResponse = {
  short_url: string;
  expires_at: string | null;
};

export type GetURLStatsAPIResponse = {
  original_url: string;
  click_count: number;
  created_at: string;
};

export type URLStats = {
  originalURL: string;
  clickCount: number;
  createdAt: Date;
};

export type ClaimURLsAPIResponse = {
  claimed_count: number;
  short_codes: string[];
};

export type ClaimURLsResponse = {
  claimedCount: number;
  shortCodes: string[];
};
