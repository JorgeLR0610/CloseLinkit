import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { shortenURL, getURLStats } from "./urls";

const baseURL = import.meta.env.VITE_API_BASE_URL || "";

const server = setupServer(
  http.post(`${baseURL}/api/v1/shorten`, async ({ request }) => {
    const body = (await request.json()) as { url: string };
    return HttpResponse.json({
      short_url: `http://localhost:8080/xyz9999`,
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
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe("urls service", () => {
  it("shortenURL sends correct payload and maps response", async () => {
    const result = await shortenURL("https://example.com/long-url");
    expect(result).toEqual({
      shortURL: "http://localhost:8080/xyz9999",
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
});
