import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { FaWhatsapp } from "react-icons/fa6";
import ShareButton from "./ShareButton";

describe("ShareButton Component", () => {
  it("renders an anchor link with correct attributes, accessibility label, and custom classes", () => {
    const testHref = "https://api.whatsapp.com/send?text=http%3A%2F%2Flocalhost%3A8080%2Fxyz1234";
    const testLabel = "WhatsApp";

    render(
      <ShareButton
        href={testHref}
        icon={FaWhatsapp}
        label={testLabel}
        className="whatsapp-custom"
      />,
    );

    const link = screen.getByRole("link", { name: `Share on ${testLabel}` });

    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", testHref);
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
    expect(link).toHaveClass("share-button");
    expect(link).toHaveClass("whatsapp-custom");

    // Verify SVG icon rendering inside anchor
    const svgIcon = link.querySelector("svg");
    expect(svgIcon).toBeInTheDocument();
  });
});
