import { describe, it, expect, beforeAll, afterAll, afterEach, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import request from "./apiClient";

const baseURL = import.meta.env.VITE_API_BASE_URL || "";

const server = setupServer(
  http.get(`${baseURL}/api/v1/test-endpoint`, () => {
    return HttpResponse.json({ success: true, message: "OK" });
  }),
  http.post(`${baseURL}/api/v1/test-json-error`, () => {
    return HttpResponse.json({ error: "Custom backend error" }, { status: 400 });
  }),
  http.get(`${baseURL}/api/v1/test-invalid-json-error`, () => {
    return new HttpResponse("Internal Server Error Plain Text", {
      status: 500,
      headers: { "Content-Type": "text/plain" },
    });
  }),
);

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  vi.restoreAllMocks();
});
afterAll(() => server.close());

describe("apiClient request utility", () => {
  it("successfully fetches and parses JSON response", async () => {
    const data = await request<{ success: boolean; message: string }>("/api/v1/test-endpoint", {
      method: "GET",
    });
    expect(data).toEqual({ success: true, message: "OK" });
  });

  it("handles backend error with custom JSON error message", async () => {
    await expect(request("/api/v1/test-json-error", { method: "POST" })).rejects.toThrow(
      "Custom backend error",
    );
  });

  it("handles backend error with non-JSON body using fallback message", async () => {
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    await expect(request("/api/v1/test-invalid-json-error", { method: "GET" })).rejects.toThrow(
      "An unexpected error occurred. Please try again later.",
    );
    consoleErrorSpy.mockRestore();
  });

  it("throws offline message on network failure when navigator.onLine is false", async () => {
    server.use(
      http.get(`${baseURL}/api/v1/network-fail`, () => {
        return HttpResponse.error();
      }),
    );

    Object.defineProperty(navigator, "onLine", {
      value: false,
      configurable: true,
      writable: true,
    });

    await expect(request("/api/v1/network-fail", { method: "GET" })).rejects.toThrow(
      "You seem to be offline. Please check your internet connection.",
    );
  });

  it("throws unreachable message on network failure when navigator.onLine is true", async () => {
    server.use(
      http.get(`${baseURL}/api/v1/network-fail`, () => {
        return HttpResponse.error();
      }),
    );

    Object.defineProperty(navigator, "onLine", {
      value: true,
      configurable: true,
      writable: true,
    });

    await expect(request("/api/v1/network-fail", { method: "GET" })).rejects.toThrow(
      "Unable to reach the service. Please try again later.",
    );
  });
});
