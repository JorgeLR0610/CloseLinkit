import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import FooterCTA from "./FooterCTA";

describe("FooterCTA Component", () => {
  it("renders footer headings, description, and action buttons", () => {
    render(<FooterCTA />);

    expect(
      screen.getByRole("heading", { level: 2, name: /want more\? sign up now!/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/custom links, powerful analytics, and much more\./i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Login" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign up" })).toBeInTheDocument();
  });
});
