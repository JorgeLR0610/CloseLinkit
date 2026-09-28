import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { MemoryRouter } from "react-router";
import toast from "react-hot-toast";
import App from "./App";
import * as urlServices from "./services/urls";
import * as AuthContextModule from "./context/useAuth";
import * as authService from "./services/auth";

function renderApp(initialEntry = "/") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <App />
    </MemoryRouter>,
  );
}

describe("App Component", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.spyOn(toast, "success").mockImplementation(() => "");
    vi.spyOn(toast, "error").mockImplementation(() => "");

    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("renders initial state without recent box or history list when localStorage is empty", () => {
    renderApp("/");

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/paste your link here/i)).toBeInTheDocument();
    expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
  });

  it("loads and displays initial history from localStorage filtering out expired items", () => {
    const futureDate = new Date(Date.now() + 1000 * 60 * 60 * 24).toISOString();
    const pastDate = new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString();

    const initialHistory = [
      {
        originalURL: "https://example.com/stored1",
        shortURL: "http://localhost:8080/s1",
        expiresAt: futureDate,
      },
      {
        originalURL: "https://example.com/expired",
        shortURL: "http://localhost:8080/s-exp",
        expiresAt: pastDate,
      },
      { originalURL: "https://example.com/no-exp", shortURL: "http://localhost:8080/s-no-exp" },
    ];
    localStorage.setItem("history", JSON.stringify(initialHistory));

    renderApp("/");

    expect(screen.getByText(/list of shortened urls/i)).toBeInTheDocument();
    expect(screen.getByText("https://example.com/stored1")).toBeInTheDocument();
    expect(screen.getByText("https://example.com/no-exp")).toBeInTheDocument();
    expect(screen.queryByText("https://example.com/expired")).not.toBeInTheDocument();
  });

  it("gracefully handles corrupt JSON in localStorage", () => {
    localStorage.setItem("history", "invalid-json{{{");

    renderApp("/");

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
  });

  it("shortens URL, displays recent box, puts new item on top of history, and saves to localStorage", async () => {
    vi.spyOn(urlServices, "shortenURL").mockResolvedValue({
      shortURL: "http://localhost:8080/new123",
      expiresAt: "2026-10-03T12:00:00Z",
    });

    renderApp("/");

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/target" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      // Recent URL Box rendered
      expect(screen.getAllByText("http://localhost:8080/new123").length).toBeGreaterThanOrEqual(1);
      // URL List rendered with original and short URL
      expect(screen.getByText("https://example.com/target")).toBeInTheDocument();
      // Ensure useEffect has flushed to localStorage
      const stored = JSON.parse(localStorage.getItem("history") || "[]");
      expect(stored).toEqual([
        {
          originalURL: "https://example.com/target",
          shortURL: "http://localhost:8080/new123",
          expiresAt: "2026-10-03T12:00:00Z",
        },
      ]);
    });
  });

  it("maintains a maximum of 10 items in history", async () => {
    const existing10Items = Array.from({ length: 10 }, (_, i) => ({
      originalURL: `https://example.com/item-${i + 1}`,
      shortURL: `http://localhost:8080/code-${i + 1}`,
    }));
    localStorage.setItem("history", JSON.stringify(existing10Items));

    vi.spyOn(urlServices, "shortenURL").mockResolvedValue({
      shortURL: "http://localhost:8080/code-11",
      expiresAt: null,
    });

    renderApp("/");

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/item-11" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText("https://example.com/item-11")).toBeInTheDocument();
      const stored = JSON.parse(localStorage.getItem("history") || "[]");
      expect(stored).toHaveLength(10);
      expect(stored[0]).toEqual({
        originalURL: "https://example.com/item-11",
        shortURL: "http://localhost:8080/code-11",
        expiresAt: null,
      });
      // 10th item is preserved, oldest (item-10) was dropped
      expect(stored[9]).toEqual({
        originalURL: "https://example.com/item-9",
        shortURL: "http://localhost:8080/code-9",
      });
    });
  });

  it("handles shortenURL failure with toast error without modifying state", async () => {
    vi.spyOn(urlServices, "shortenURL").mockRejectedValue(new Error("Network error on API"));

    renderApp("/");

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/failed" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Network error on API");
      expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
    });
  });

  it("routes to LoginPage on /login", () => {
    renderApp("/login");

    expect(screen.getByRole("heading", { name: "Welcome Back" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Log In" })).toBeInTheDocument();
  });

  it("routes to SignupPage on /signup", () => {
    renderApp("/signup");

    expect(screen.getByRole("heading", { name: "Create Account" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign Up" })).toBeInTheDocument();
  });

  it("redirects from /dashboard to /", () => {
    renderApp("/dashboard");

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/paste your link here/i)).toBeInTheDocument();
  });

  it("loads and displays authenticated user URLs from API on /", async () => {
    const mockURLs = [
      {
        original_url: "https://example.com/user-link-1",
        short_code: "code1",
        short_url: "http://localhost:8080/code1",
        created_at: "2026-09-20T10:00:00Z",
        click_count: 5,
      },
    ];

    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: { id: "123", email: "user@example.com", created_at: "" },
      accessToken: "user-jwt-token",
      isLoading: false,
      isAuthenticated: true,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });
    vi.spyOn(authService, "getUserURLs").mockResolvedValue(mockURLs);

    renderApp("/");

    await waitFor(() => {
      expect(authService.getUserURLs).toHaveBeenCalledWith("user-jwt-token");
      expect(screen.getByText("https://example.com/user-link-1")).toBeInTheDocument();
      expect(screen.getByText("http://localhost:8080/code1")).toBeInTheDocument();
    });
  });

  it("allows authenticated user to shorten URL, calls API with token, and updates user URLs", async () => {
    const initialURLs = [
      {
        original_url: "https://example.com/first",
        short_code: "first1",
        short_url: "http://localhost:8080/first1",
        created_at: "2026-09-20T10:00:00Z",
        click_count: 0,
      },
    ];
    const updatedURLs = [
      {
        original_url: "https://example.com/second",
        short_code: "second2",
        short_url: "http://localhost:8080/second2",
        created_at: "2026-09-27T10:00:00Z",
        click_count: 0,
      },
      ...initialURLs,
    ];

    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: { id: "123", email: "user@example.com", created_at: "" },
      accessToken: "user-jwt-token",
      isLoading: false,
      isAuthenticated: true,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    const getUserURLsSpy = vi
      .spyOn(authService, "getUserURLs")
      .mockResolvedValueOnce(initialURLs)
      .mockResolvedValueOnce(updatedURLs);

    vi.spyOn(urlServices, "shortenURL").mockResolvedValue({
      shortURL: "http://localhost:8080/second2",
      expiresAt: null,
    });

    renderApp("/");

    await waitFor(() => {
      expect(screen.getByText("https://example.com/first")).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/second" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(urlServices.shortenURL).toHaveBeenCalledWith(
        "https://example.com/second",
        "user-jwt-token",
      );
      expect(getUserURLsSpy).toHaveBeenCalledTimes(2);
      expect(screen.getByText("https://example.com/second")).toBeInTheDocument();
      expect(localStorage.getItem("history")).toBeNull();
    });
  });

  it("redirects unknown path to home /", () => {
    renderApp("/some-nonexistent-route");

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/paste your link here/i)).toBeInTheDocument();
  });
});
