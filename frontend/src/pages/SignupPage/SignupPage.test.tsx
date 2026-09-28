import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { MemoryRouter, Routes, Route } from "react-router";
import SignupPage from "./SignupPage";
import * as AuthContextModule from "../../context/useAuth";
import toast from "react-hot-toast";

describe("SignupPage Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(toast, "success").mockImplementation(() => "");
    vi.spyOn(toast, "error").mockImplementation(() => "");
  });

  it("renders signup form inputs, labels, and links", () => {
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
        <SignupPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Create Account" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByLabelText("Confirm Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign Up" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Log in" })).toHaveAttribute("href", "/login");
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
        <SignupPage />
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

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
        <SignupPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "not-an-email" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

    expect(await screen.findByText("Please enter a valid email address.")).toBeInTheDocument();
  });

  it("validates password length under 8 characters", async () => {
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
        <SignupPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "short" } });
    fireEvent.change(screen.getByLabelText("Confirm Password"), { target: { value: "short" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

    expect(
      await screen.findByText("Password must be at least 8 characters long."),
    ).toBeInTheDocument();
  });

  it("validates password mismatch", async () => {
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
        <SignupPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password123" } });
    fireEvent.change(screen.getByLabelText("Confirm Password"), {
      target: { value: "different123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

    expect(await screen.findByText("Passwords do not match.")).toBeInTheDocument();
  });

  it("submits valid registration and redirects to home with toast", async () => {
    const mockRegister = vi.fn().mockResolvedValue(undefined);
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: mockRegister,
      logout: vi.fn(),
    });

    render(
      <MemoryRouter initialEntries={["/signup"]}>
        <Routes>
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/" element={<div>Home Content</div>} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "user@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password123" } });
    fireEvent.change(screen.getByLabelText("Confirm Password"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith("user@example.com", "password123");
      expect(toast.success).toHaveBeenCalledWith(
        "Welcome to CloseLinkit! Your account has been created.",
      );
      expect(screen.getByText("Home Content")).toBeInTheDocument();
    });
  });

  it("displays error message on failed registration", async () => {
    const mockRegister = vi
      .fn()
      .mockRejectedValue(new Error("user with this email already exists"));
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: null,
      accessToken: null,
      isLoading: false,
      isAuthenticated: false,
      login: vi.fn(),
      register: mockRegister,
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <SignupPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "existing@example.com" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password123" } });
    fireEvent.change(screen.getByLabelText("Confirm Password"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Sign Up" }));

    await waitFor(() => {
      expect(screen.getByText("user with this email already exists")).toBeInTheDocument();
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
      <MemoryRouter initialEntries={["/signup"]}>
        <Routes>
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/" element={<div>Home View</div>} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Home View")).toBeInTheDocument();
  });
});
