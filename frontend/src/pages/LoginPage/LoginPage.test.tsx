import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { MemoryRouter, Routes, Route } from "react-router";
import LoginPage from "./LoginPage";
import * as AuthContextModule from "../../context/useAuth";
import toast from "react-hot-toast";

describe("LoginPage Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(toast, "success").mockImplementation(() => "");
    vi.spyOn(toast, "error").mockImplementation(() => "");
  });

  it("renders login form inputs, labels, and links", () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Welcome Back" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Log In" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Sign up" })).toHaveAttribute("href", "/signup");
    expect(screen.getByRole("link", { name: "← Back to Home" })).toHaveAttribute("href", "/");
  });

  it("validates empty email", async () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Log In" }));

    expect(await screen.findByText("Email address is required.")).toBeInTheDocument();
  });

  it("validates invalid email format", async () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "invalid-email" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password123" } });
    fireEvent.click(screen.getByRole("button", { name: "Log In" }));

    expect(await screen.findByText("Please enter a valid email address.")).toBeInTheDocument();
  });

  it("validates missing password", async () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.click(screen.getByRole("button", { name: "Log In" }));

    expect(await screen.findByText("Password is required.")).toBeInTheDocument();
  });

  it("submits valid credentials and logs in successfully", async () => {
    const mockLogin = vi.fn().mockResolvedValue(undefined);
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: mockLogin,
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/" element={<div>Home Content</div>} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password123" } });
    fireEvent.click(screen.getByRole("button", { name: "Log In" }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith("user@example.com", "password123");
      expect(toast.success).toHaveBeenCalledWith("Welcome back!");
      expect(screen.getByText("Home Content")).toBeInTheDocument();
    });
  });

  it("displays error message on failed login", async () => {
    const mockLogin = vi.fn().mockRejectedValue(new Error("invalid email or password"));
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: mockLogin,
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrongpassword" } });
    fireEvent.click(screen.getByRole("button", { name: "Log In" }));

    await waitFor(() => {
      expect(screen.getByText("invalid email or password")).toBeInTheDocument();
    });
  });

  it("redirects to home when already authenticated", () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: { id: "123", email: "user@example.com", created_at: "" },
      accessToken: "token",
      isLoading: false,
      isAuthenticated: true,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/" element={<div>Home View</div>} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Home View")).toBeInTheDocument();
  });
});
