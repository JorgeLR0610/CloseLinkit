import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import RecentURLBox from "./RecentURLBox";

describe("RecentURLBox Component", () => {
  const shortURL = "http://localhost:8080/abc1234";

  it("renders shortURL text, CopyButton, and social share links with encoded URL", () => {
    render(<RecentURLBox shortURL={shortURL} />);

    // Check rendered short URL
    expect(screen.getByText(shortURL)).toBeInTheDocument();

    // Check CopyButton presence
    const copyButton = screen.getByRole("button", { name: "Copy" });
    expect(copyButton).toBeInTheDocument();

    const encodedURL = encodeURIComponent(shortURL);

    // Check WhatsApp share link
    const whatsapp = screen.getByRole("link", { name: "Share on WhatsApp" });
    expect(whatsapp).toBeInTheDocument();
    expect(whatsapp).toHaveAttribute("href", `https://api.whatsapp.com/send?text=${encodedURL}`);

    // Check Facebook share link
    const facebook = screen.getByRole("link", { name: "Share on facebook" });
    expect(facebook).toBeInTheDocument();
    expect(facebook).toHaveAttribute(
      "href",
      `https://www.facebook.com/sharer/sharer.php?u=${encodedURL}`,
    );

    // Check LinkedIn share link
    const linkedin = screen.getByRole("link", { name: "Share on LinkedIn" });
    expect(linkedin).toBeInTheDocument();
    expect(linkedin).toHaveAttribute(
      "href",
      `https://www.linkedin.com/sharing/share-offsite/?url=${encodedURL}`,
    );

    // Check Telegram share link
    const telegram = screen.getByRole("link", { name: "Share on Telegram" });
    expect(telegram).toBeInTheDocument();
    expect(telegram).toHaveAttribute("href", `https://t.me/share/url?url=${encodedURL}`);

    // Check X / Twitter share link
    const xTwitter = screen.getByRole("link", { name: "Share on X" });
    expect(xTwitter).toBeInTheDocument();
    expect(xTwitter).toHaveAttribute(
      "href",
      `https://x.com/intent/post?text=${encodeURIComponent("Check out this link!")}&url=${encodedURL}`,
    );
  });
});
