import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, afterEach } from "vitest";
import URLListItem from "./URLListItem";
import * as urlServices from "../../services/urls";

describe("URLListItem Component", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

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

  const itemWithStats = {
    originalURL: "https://example.com/preloaded",
    shortURL: "http://localhost:8080/preload1",
    clickCount: 42,
    createdAt: "2026-09-01T12:00:00Z",
  };

  it("refreshes pre-populated stats from the API when the panel is opened", async () => {
    const getStatsSpy = vi.spyOn(urlServices, "getURLStats").mockResolvedValue({
      originalURL: itemWithStats.originalURL,
      clickCount: 50,
      createdAt: new Date("2026-09-01T12:00:00Z"),
    });

    render(<URLListItem item={itemWithStats} />);
    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));

    expect(await screen.findByText("50")).toBeInTheDocument();
    expect(screen.queryByText("42")).not.toBeInTheDocument();
    expect(getStatsSpy).toHaveBeenCalledTimes(1);
  });

  it("shows pre-populated stats when the refresh request fails", async () => {
    vi.spyOn(urlServices, "getURLStats").mockRejectedValue(new Error("network"));

    render(<URLListItem item={itemWithStats} />);
    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));

    expect(await screen.findByText("42")).toBeInTheDocument();
  });

  it("does not open the panel when there are no stats and the request fails", async () => {
    const getStatsSpy = vi
      .spyOn(urlServices, "getURLStats")
      .mockRejectedValue(new Error("network"));

    render(<URLListItem item={item} />);
    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));

    await waitFor(() => expect(getStatsSpy).toHaveBeenCalledTimes(1));
    expect(screen.queryByText("Total Clicks")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Analytics" })).toBeInTheDocument();
  });

  it("requests fresh stats every time the panel is opened", async () => {
    const getStatsSpy = vi
      .spyOn(urlServices, "getURLStats")
      .mockResolvedValueOnce({
        originalURL: item.originalURL,
        clickCount: 1,
        createdAt: new Date("2026-08-10T15:30:00Z"),
      })
      .mockResolvedValueOnce({
        originalURL: item.originalURL,
        clickCount: 2,
        createdAt: new Date("2026-08-10T15:30:00Z"),
      });

    render(<URLListItem item={item} />);

    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));
    expect(await screen.findByText("1")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Shrink" }));
    expect(screen.queryByText("Total Clicks")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Analytics" }));
    expect(await screen.findByText("2")).toBeInTheDocument();
    expect(getStatsSpy).toHaveBeenCalledTimes(2);
  });

  it("does not render Delete button when onDelete is not provided", () => {
    render(<URLListItem item={item} />);
    expect(screen.queryByRole("button", { name: /delete/i })).not.toBeInTheDocument();
  });

  it("renders Delete button when onDelete is provided and calls onDelete on click", async () => {
    const onDeleteMock = vi.fn().mockResolvedValue(undefined);
    render(<URLListItem item={item} onDelete={onDeleteMock} />);

    const deleteBtn = screen.getByRole("button", { name: /delete/i });
    expect(deleteBtn).toBeInTheDocument();

    fireEvent.click(deleteBtn);
    expect(onDeleteMock).toHaveBeenCalledWith(item.shortURL);
    await waitFor(() => {
      expect(deleteBtn).not.toBeDisabled();
    });
  });

  it("disables Delete button while deletion is pending", async () => {
    let resolveDelete: () => void = () => {};
    const pendingPromise = new Promise<void>((resolve) => {
      resolveDelete = resolve;
    });
    const onDeleteMock = vi.fn().mockReturnValue(pendingPromise);

    render(<URLListItem item={item} onDelete={onDeleteMock} />);
    const deleteBtn = screen.getByRole("button", { name: /delete/i });

    fireEvent.click(deleteBtn);
    expect(deleteBtn).toBeDisabled();
    expect(screen.getByRole("button", { name: /deleting/i })).toBeInTheDocument();

    resolveDelete();
    await waitFor(() => {
      expect(deleteBtn).not.toBeDisabled();
    });
  });
});
