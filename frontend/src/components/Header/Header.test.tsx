import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import Header from "./Header";

describe("Header Component", () => {
  it("renders logo image and brand name", () => {
    render(<Header />);

    const logoImg = screen.getByAltText("CloseLinkit");
    expect(logoImg).toBeInTheDocument();

    const brandText = screen.getByText("CloseLinkit");
    expect(brandText).toBeInTheDocument();
  });
});
