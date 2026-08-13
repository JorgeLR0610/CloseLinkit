export type URLItem = {
  originalURL: string;
  shortURL: string;
};

export type ShortenURLResponse = {
  shortURL: string;
};

export type ShortenURLAPIResponse = {
  short_url: string;
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
