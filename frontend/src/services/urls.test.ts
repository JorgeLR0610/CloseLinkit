import { describe, it, expect, beforeAll, afterAll, afterEach, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { shortenURL, getURLStats, claimURLs, claimStoredURLs, deleteURL } from "./urls";

const baseURL = import.meta.env.VITE_API_BASE_URL || "";

const server = setupServer(
  http.post(`${baseURL}/api/v1/shorten`, async ({ request }) => {
    const body = (await request.json()) as { url: string };
    return HttpResponse.json({
      short_url: `http://localhost:8080/xyz9999`,
      expires_at: "2026-10-03T12:00:00Z",
      original_url: body.url,
    });
  }),
  http.get(`${baseURL}/api/v1/xyz9999/stats`, () => {
    return HttpResponse.json({
      original_url: "https://example.com/test-stats",
      click_count: 42,
      created_at: "2026-08-01T12:00:00Z",
    });
  }),
  http.post(`${baseURL}/api/v1/urls/claim`, async ({ request }) => {
    const body = (await request.json()) as { short_codes: string[] };
    return HttpResponse.json({
      claimed_count: body.short_codes.length,
      short_codes: body.short_codes,
    });
  }),
);

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  localStorage.clear();
  vi.restoreAllMocks();
});
afterAll(() => server.close());

describe("urls service", () => {
  it("shortenURL sends correct payload and maps response including expiresAt", async () => {
    const result = await shortenURL("https://example.com/long-url");
    expect(result).toEqual({
      shortURL: "http://localhost:8080/xyz9999",
      expiresAt: "2026-10-03T12:00:00Z",
    });
  });

  it("shortenURL attaches Authorization header when accessToken is provided", async () => {
    let capturedAuthHeader: string | null = null;
    server.use(
      http.post(`${baseURL}/api/v1/shorten`, ({ request }) => {
        capturedAuthHeader = request.headers.get("Authorization");
        return HttpResponse.json({
          short_url: `http://localhost:8080/perm123`,
          expires_at: null,
        });
      }),
    );

    const result = await shortenURL("https://example.com/perm", "my-jwt-token");
    expect(capturedAuthHeader).toBe("Bearer my-jwt-token");
    expect(result).toEqual({
      shortURL: "http://localhost:8080/perm123",
      expiresAt: null,
    });
  });

  it("getURLStats extracts shortCode from full shortURL and converts created_at to Date", async () => {
    const stats = await getURLStats("http://localhost:8080/xyz9999");
    expect(stats).toEqual({
      originalURL: "https://example.com/test-stats",
      clickCount: 42,
      createdAt: new Date("2026-08-01T12:00:00Z"),
    });
    expect(stats.createdAt).toBeInstanceOf(Date);
  });

  it("claimURLs sends Authorization header and payload, and maps response", async () => {
    let capturedAuthHeader: string | null = null;
    let capturedBody: unknown = null;

    server.use(
      http.post(`${baseURL}/api/v1/urls/claim`, async ({ request }) => {
        capturedAuthHeader = request.headers.get("Authorization");
        capturedBody = await request.json();
        return HttpResponse.json({
          claimed_count: 2,
          short_codes: ["code1", "code2"],
        });
      }),
    );

    const result = await claimURLs(["code1", "code2"], "claim-token");
    expect(capturedAuthHeader).toBe("Bearer claim-token");
    expect(capturedBody).toEqual({ short_codes: ["code1", "code2"] });
    expect(result).toEqual({
      claimedCount: 2,
      shortCodes: ["code1", "code2"],
    });
  });

  it("claimStoredURLs returns null when no history in localStorage", async () => {
    const result = await claimStoredURLs("claim-token");
    expect(result).toBeNull();
  });

  it("claimStoredURLs returns null when localStorage has invalid json", async () => {
    localStorage.setItem("history", "invalid-json{{");
    const result = await claimStoredURLs("claim-token");
    expect(result).toBeNull();
  });

  it("claimStoredURLs claims valid items and clears localStorage upon success", async () => {
    const futureDate = new Date(Date.now() + 1000 * 60 * 60 * 24).toISOString();
    const pastDate = new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString();

    const storedItems = [
      {
        originalURL: "https://example.com/valid1",
        shortURL: "http://localhost:8080/code1",
        expiresAt: futureDate,
      },
      {
        originalURL: "https://example.com/expired",
        shortURL: "http://localhost:8080/exp-code",
        expiresAt: pastDate,
      },
      { originalURL: "https://example.com/valid2", shortURL: "http://localhost:8080/code2" },
    ];
    localStorage.setItem("history", JSON.stringify(storedItems));

    const result = await claimStoredURLs("claim-token");
    expect(result).toEqual({
      claimedCount: 2,
      shortCodes: ["code1", "code2"],
    });
    expect(localStorage.getItem("history")).toBeNull();
  });

  it("claimStoredURLs removes history and returns null if all items are expired", async () => {
    const pastDate = new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString();
    const storedItems = [
      {
        originalURL: "https://example.com/expired",
        shortURL: "http://localhost:8080/exp-code",
        expiresAt: pastDate,
      },
    ];
    localStorage.setItem("history", JSON.stringify(storedItems));

    const result = await claimStoredURLs("claim-token");
    expect(result).toBeNull();
    expect(localStorage.getItem("history")).toBeNull();
  });

  it("claimStoredURLs preserves localStorage history when claim request fails", async () => {
    server.use(
      http.post(`${baseURL}/api/v1/urls/claim`, () => {
        return HttpResponse.json({ error: "Server error" }, { status: 500 });
      }),
    );

    const storedItems = [
      { originalURL: "https://example.com/preserve", shortURL: "http://localhost:8080/pres1" },
    ];
    localStorage.setItem("history", JSON.stringify(storedItems));

    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await claimStoredURLs("claim-token");
    expect(result).toBeNull();
    expect(localStorage.getItem("history")).not.toBeNull();
    consoleSpy.mockRestore();
  });

  it("deleteURL sends DELETE request with Authorization header and handles 204 No Content", async () => {
    let capturedAuthHeader: string | null = null;
    let capturedCode: string | null = null;

    server.use(
      http.delete(`${baseURL}/api/v1/urls/:shortCode`, ({ request, params }) => {
        capturedAuthHeader = request.headers.get("Authorization");
        capturedCode = params.shortCode as string;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await expect(deleteURL("myCode123", "test-access-token")).resolves.toBeUndefined();
    expect(capturedAuthHeader).toBe("Bearer test-access-token");
    expect(capturedCode).toBe("myCode123");
  });

  it("deleteURL extracts short code from full short URL", async () => {
    let capturedCode: string | null = null;

    server.use(
      http.delete(`${baseURL}/api/v1/urls/:shortCode`, ({ params }) => {
        capturedCode = params.shortCode as string;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await deleteURL("http://localhost:8080/extractedCode", "test-access-token");
    expect(capturedCode).toBe("extractedCode");
  });

  it("deleteURL throws error on 404 Not Found response", async () => {
    server.use(
      http.delete(`${baseURL}/api/v1/urls/:shortCode`, () => {
        return HttpResponse.json({ error: "Not found" }, { status: 404 });
      }),
    );

    await expect(deleteURL("missingCode", "test-access-token")).rejects.toThrow("Not found");
  });

  it("deleteURL throws error on 401 Unauthorized response", async () => {
    server.use(
      http.delete(`${baseURL}/api/v1/urls/:shortCode`, () => {
        return HttpResponse.json({ error: "Unauthorized" }, { status: 401 });
      }),
    );

    await expect(deleteURL("anyCode", "invalid-token")).rejects.toThrow("Unauthorized");
  });
});
