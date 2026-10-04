import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { register, login, refreshToken, logout, getUserURLs } from "./auth";

const baseURL = import.meta.env.VITE_API_BASE_URL || "";

const mockUser = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  email: "test@example.com",
  created_at: "2026-09-20T10:00:00Z",
};

const server = setupServer(
  http.post(`${baseURL}/api/v1/auth/register`, async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string };
    if (body.email === "existing@example.com") {
      return HttpResponse.json({ error: "user with this email already exists" }, { status: 409 });
    }
    return HttpResponse.json({ ...mockUser, email: body.email }, { status: 201 });
  }),

  http.post(`${baseURL}/api/v1/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string };
    if (body.password === "wrongpassword") {
      return HttpResponse.json({ error: "invalid email or password" }, { status: 401 });
    }
    return HttpResponse.json({
      access_token: "mock-access-token",
      expires_in: 900,
      user: { ...mockUser, email: body.email },
    });
  }),

  http.post(`${baseURL}/api/v1/auth/refresh`, () => {
    return HttpResponse.json({
      access_token: "new-access-token",
      expires_in: 900,
      user: mockUser,
    });
  }),

  http.post(`${baseURL}/api/v1/auth/logout`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${baseURL}/api/v1/urls`, ({ request }) => {
    const authHeader = request.headers.get("Authorization");
    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      return HttpResponse.json({ error: "unauthorized" }, { status: 401 });
    }
    return HttpResponse.json([
      {
        original_url: "https://example.com/dest",
        short_code: "abc1234",
        short_url: "http://localhost:8080/abc1234",
        created_at: "2026-09-21T12:00:00Z",
        click_count: 5,
      },
    ]);
  }),
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe("auth service", () => {
  it("register successfully registers a new user", async () => {
    const res = await register({ email: "new@example.com", password: "password123" });
    expect(res.email).toBe("new@example.com");
    expect(res.id).toBe(mockUser.id);
  });

  it("register throws when email is already taken", async () => {
    await expect(
      register({ email: "existing@example.com", password: "password123" }),
    ).rejects.toThrow("user with this email already exists");
  });

  it("login successfully returns tokens and user", async () => {
    const res = await login({ email: "test@example.com", password: "password123" });
    expect(res.access_token).toBe("mock-access-token");
    expect(res.user.email).toBe("test@example.com");
  });

  it("login throws on invalid credentials", async () => {
    await expect(login({ email: "test@example.com", password: "wrongpassword" })).rejects.toThrow(
      "invalid email or password",
    );
  });

  it("refreshToken successfully requests new token without body", async () => {
    const res = await refreshToken();
    expect(res.access_token).toBe("new-access-token");
    expect(res.user.email).toBe(mockUser.email);
  });

  it("logout successfully requests logout endpoint with 204", async () => {
    await expect(logout()).resolves.toBeUndefined();
  });

  it("getUserURLs returns list of urls when authenticated", async () => {
    const urls = await getUserURLs("valid-token");
    expect(urls).toHaveLength(1);
    expect(urls[0].short_code).toBe("abc1234");
    expect(urls[0].click_count).toBe(5);
  });

  it("getUserURLs throws when unauthorized", async () => {
    server.use(
      http.get(`${baseURL}/api/v1/urls`, () => {
        return HttpResponse.json({ error: "unauthorized" }, { status: 401 });
      }),
    );
    await expect(getUserURLs("invalid-token")).rejects.toThrow("unauthorized");
  });
});
