import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MemoryRouter } from "react-router";
import FooterCTA from "./FooterCTA";
import * as AuthContextModule from "../../context/useAuth";

describe("FooterCTA Component", () => {
  it("renders guest headings, description, and action links when unauthenticated", () => {
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
        <FooterCTA />
      </MemoryRouter>,
    );

    expect(
      screen.getByRole("heading", { level: 2, name: /want more\? sign up now!/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/custom links, powerful analytics, and much more\./i),
    ).toBeInTheDocument();

    const loginLink = screen.getByRole("link", { name: "Login" });
    expect(loginLink).toBeInTheDocument();
    expect(loginLink).toHaveAttribute("href", "/login");

    const signupLink = screen.getByRole("link", { name: "Sign up" });
    expect(signupLink).toBeInTheDocument();
    expect(signupLink).toHaveAttribute("href", "/signup");
  });

  it("renders nothing when authenticated", () => {
    vi.spyOn(AuthContextModule, "useAuth").mockReturnValue({
      user: { id: "123", email: "user@example.com", created_at: "" },
      accessToken: "token",
      isLoading: false,
      isAuthenticated: true,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
    });

    const { container } = render(
      <MemoryRouter>
        <FooterCTA />
      </MemoryRouter>,
    );

    expect(container.firstChild).toBeNull();
  });
});
