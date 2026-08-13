import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import URLListItem from "./URLListItem";
import * as urlServices from "../../services/urls";

describe("URLListItem Component", () => {
  const item = {
    originalURL: "https://example.com/very-long-original-url",
    shortURL: "http://localhost:8080/xyz1234",
  };

  it("renders original URL, short URL link, Copy button, and Stats button", () => {
    render(<URLListItem item={item} />);

    expect(screen.getByText(item.originalURL)).toBeInTheDocument();

    const shortUrlLink = screen.getByRole("link", { name: item.shortURL });
    expect(shortUrlLink).toBeInTheDocument();
    expect(shortUrlLink).toHaveAttribute("href", item.shortURL);

    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Analytics" })).toBeInTheDocument();
  });

  it("displays analytics stats subpanel when Analytics button is clicked", async () => {
    const mockDate = new Date("2026-08-10T15:30:00Z");
    vi.spyOn(urlServices, "getURLStats").mockResolvedValue({
      originalURL: item.originalURL,
      clickCount: 88,
      createdAt: mockDate,
    });

    render(<URLListItem item={item} />);

    const analyticsButton = screen.getByRole("button", { name: "Analytics" });
    fireEvent.click(analyticsButton);

    await waitFor(() => {
      expect(screen.getByText("Total Clicks")).toBeInTheDocument();
      expect(screen.getByText("88")).toBeInTheDocument();
      expect(screen.getByText("Created On")).toBeInTheDocument();
    });
  });
});
