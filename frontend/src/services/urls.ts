import type {
  GetURLStatsAPIResponse,
  URLStats,
  ShortenURLAPIResponse,
  ShortenURLResponse,
  ClaimURLsAPIResponse,
  ClaimURLsResponse,
  URLItem,
} from "../types/url";
import request from "./apiClient";

export async function shortenURL(
  originalURL: string,
  accessToken?: string,
): Promise<ShortenURLResponse> {
  const headers: Record<string, string> = {};
  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const data = await request<ShortenURLAPIResponse>("/api/v1/shorten", {
    method: "POST",
    headers,
    body: JSON.stringify({
      url: originalURL,
    }),
  });

  return {
    shortURL: data.short_url,
    expiresAt: data.expires_at,
  };
}

export async function getURLStats(shortURL: string): Promise<URLStats> {
  const shortCode = shortURL.split("/").pop();

  const data = await request<GetURLStatsAPIResponse>(`/api/v1/${shortCode}/stats`, {
    method: "GET",
  });

  return {
    originalURL: data.original_url,
    clickCount: data.click_count,
    createdAt: new Date(data.created_at),
  };
}

export async function claimURLs(
  shortCodes: string[],
  accessToken: string,
): Promise<ClaimURLsResponse> {
  const data = await request<ClaimURLsAPIResponse>("/api/v1/urls/claim", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({
      short_codes: shortCodes,
    }),
  });

  return {
    claimedCount: data.claimed_count,
    shortCodes: data.short_codes,
  };
}

export async function claimStoredURLs(accessToken: string): Promise<ClaimURLsResponse | null> {
  const stored = localStorage.getItem("history");
  if (!stored) return null;

  let parsed: URLItem[];
  try {
    parsed = JSON.parse(stored);
  } catch {
    return null;
  }

  if (!Array.isArray(parsed) || parsed.length === 0) return null;

  const now = Date.now();
  const validItems = parsed.filter(
    (item) => !item.expiresAt || new Date(item.expiresAt).getTime() > now,
  );

  if (validItems.length === 0) {
    localStorage.removeItem("history");
    return null;
  }

  const shortCodes = validItems
    .map((item) => item.shortURL.split("/").pop() || "")
    .filter((code) => code.length > 0);

  if (shortCodes.length === 0) {
    localStorage.removeItem("history");
    return null;
  }

  try {
    const res = await claimURLs(shortCodes, accessToken);
    localStorage.removeItem("history");
    return res;
  } catch (error) {
    console.error("Failed to claim stored URLs:", error);
    return null;
  }
}
