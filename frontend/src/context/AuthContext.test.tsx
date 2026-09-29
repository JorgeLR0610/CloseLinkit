import { render, screen, waitFor, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { AuthProvider } from "./AuthProvider";
import { useAuth } from "./useAuth";
import * as authService from "../services/auth";
import * as urlService from "../services/urls";

const mockUser = {
  id: "user-123",
  email: "test@example.com",
  created_at: "2026-09-20T10:00:00Z",
};

function TestConsumer() {
  const { user, accessToken, isLoading, isAuthenticated, login, register, logout } = useAuth();

  if (isLoading) return <div>Loading session...</div>;

  return (
    <div>
      <div data-testid="auth-state">{isAuthenticated ? "authenticated" : "guest"}</div>
      <div data-testid="user-email">{user?.email || "none"}</div>
      <div data-testid="token">{accessToken || "none"}</div>
      <button onClick={() => login("login@example.com", "pass123")}>Do Login</button>
      <button onClick={() => register("signup@example.com", "pass123")}>Do Register</button>
      <button onClick={() => logout()}>Do Logout</button>
    </div>
  );
}

describe("AuthContext", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("throws error when useAuth is used outside AuthProvider", () => {
    // Suppress console.error for expected error boundary log
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    expect(() => render(<TestConsumer />)).toThrow("useAuth must be used within an AuthProvider");
    consoleSpy.mockRestore();
  });

  it("restores session on mount when cookie refresh succeeds and calls claimStoredURLs", async () => {
    const claimSpy = vi.spyOn(urlService, "claimStoredURLs").mockResolvedValue(null);
    vi.spyOn(authService, "refreshToken").mockResolvedValue({
      access_token: "restored-token",
      expires_in: 900,
      user: mockUser,
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    expect(screen.getByText("Loading session...")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("authenticated");
      expect(screen.getByTestId("user-email")).toHaveTextContent("test@example.com");
      expect(screen.getByTestId("token")).toHaveTextContent("restored-token");
      expect(claimSpy).toHaveBeenCalledWith("restored-token");
    });
  });

  it("sets unauthenticated state when cookie refresh fails on mount", async () => {
    vi.spyOn(authService, "refreshToken").mockRejectedValue(new Error("Unauthorized"));

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("guest");
      expect(screen.getByTestId("user-email")).toHaveTextContent("none");
      expect(screen.getByTestId("token")).toHaveTextContent("none");
    });
  });

  it("logs in successfully, claims stored URLs, and updates state", async () => {
    const claimSpy = vi.spyOn(urlService, "claimStoredURLs").mockResolvedValue({
      claimedCount: 2,
      shortCodes: ["c1", "c2"],
    });
    vi.spyOn(authService, "refreshToken").mockRejectedValue(new Error("No session"));
    vi.spyOn(authService, "login").mockResolvedValue({
      access_token: "login-token",
      expires_in: 900,
      user: { ...mockUser, email: "login@example.com" },
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("guest");
    });

    await act(async () => {
      screen.getByText("Do Login").click();
    });

    await waitFor(() => {
      expect(claimSpy).toHaveBeenCalledWith("login-token");
      expect(screen.getByTestId("auth-state")).toHaveTextContent("authenticated");
      expect(screen.getByTestId("user-email")).toHaveTextContent("login@example.com");
      expect(screen.getByTestId("token")).toHaveTextContent("login-token");
    });
  });

  it("login succeeds even if claimStoredURLs fails", async () => {
    vi.spyOn(urlService, "claimStoredURLs").mockRejectedValue(new Error("Claim network error"));
    vi.spyOn(authService, "refreshToken").mockRejectedValue(new Error("No session"));
    vi.spyOn(authService, "login").mockResolvedValue({
      access_token: "login-token",
      expires_in: 900,
      user: { ...mockUser, email: "login@example.com" },
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("guest");
    });

    await act(async () => {
      screen.getByText("Do Login").click();
    });

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("authenticated");
      expect(screen.getByTestId("token")).toHaveTextContent("login-token");
    });
  });

  it("registers and auto-logins successfully, invoking claimStoredURLs", async () => {
    const claimSpy = vi.spyOn(urlService, "claimStoredURLs").mockResolvedValue(null);
    vi.spyOn(authService, "refreshToken").mockRejectedValue(new Error("No session"));
    vi.spyOn(authService, "register").mockResolvedValue({
      ...mockUser,
      email: "signup@example.com",
    });
    vi.spyOn(authService, "login").mockResolvedValue({
      access_token: "auto-login-token",
      expires_in: 900,
      user: { ...mockUser, email: "signup@example.com" },
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("guest");
    });

    await act(async () => {
      screen.getByText("Do Register").click();
    });

    await waitFor(() => {
      expect(claimSpy).toHaveBeenCalledWith("auto-login-token");
      expect(screen.getByTestId("auth-state")).toHaveTextContent("authenticated");
      expect(screen.getByTestId("user-email")).toHaveTextContent("signup@example.com");
      expect(screen.getByTestId("token")).toHaveTextContent("auto-login-token");
    });
  });

  it("logs out and clears state", async () => {
    vi.spyOn(authService, "refreshToken").mockResolvedValue({
      access_token: "restored-token",
      expires_in: 900,
      user: mockUser,
    });
    vi.spyOn(authService, "logout").mockResolvedValue();

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("authenticated");
    });

    await act(async () => {
      screen.getByText("Do Logout").click();
    });

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("guest");
      expect(screen.getByTestId("user-email")).toHaveTextContent("none");
      expect(screen.getByTestId("token")).toHaveTextContent("none");
    });
  });
});
