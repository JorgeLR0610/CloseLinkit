import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import toast from "react-hot-toast";
import StatsButton from "./StatsButton";
import * as urlServices from "../../services/urls";

describe("StatsButton Component", () => {
  const shortURL = "http://localhost:8080/xyz1234";
  const setDisplayedStatsMock = vi.fn();
  const setStatsMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(toast, "error").mockImplementation(() => "");
  });

  it('renders initial button state with text "Analytics"', () => {
    render(
      <StatsButton
        shortURL={shortURL}
        displayedStats={false}
        setDisplayedStats={setDisplayedStatsMock}
        setStats={setStatsMock}
      />,
    );

    expect(screen.getByRole("button", { name: "Analytics" })).toBeInTheDocument();
  });

  it('fetches stats on click, updates state, and toggles to "Shrink"', async () => {
    const mockStats = {
      originalURL: "https://example.com",
      clickCount: 15,
      createdAt: new Date("2026-08-01T10:00:00Z"),
    };
    vi.spyOn(urlServices, "getURLStats").mockResolvedValue(mockStats);

    render(
      <StatsButton
        shortURL={shortURL}
        displayedStats={false}
        setDisplayedStats={setDisplayedStatsMock}
        setStats={setStatsMock}
      />,
    );

    const button = screen.getByRole("button", { name: "Analytics" });
    fireEvent.click(button);

    await waitFor(() => {
      expect(urlServices.getURLStats).toHaveBeenCalledWith(shortURL);
      expect(setStatsMock).toHaveBeenCalledWith(mockStats);
      expect(setDisplayedStatsMock).toHaveBeenCalled();
    });
  });

  it("shows error toast and does not toggle when API fails", async () => {
    vi.spyOn(urlServices, "getURLStats").mockRejectedValue(new Error("Network error"));

    render(
      <StatsButton
        shortURL={shortURL}
        displayedStats={false}
        setDisplayedStats={setDisplayedStatsMock}
        setStats={setStatsMock}
      />,
    );

    const button = screen.getByRole("button", { name: "Analytics" });
    fireEvent.click(button);

    await waitFor(() => {
      expect(urlServices.getURLStats).toHaveBeenCalledWith(shortURL);
      expect(toast.error).toHaveBeenCalledWith("Could not load stats. Please try again later.");
      expect(setDisplayedStatsMock).not.toHaveBeenCalled();
      expect(button).toHaveTextContent("Analytics");
    });
  });

  it('renders "Shrink" when displayedStats is true and collapses on click without re-fetching', () => {
    const getStatsSpy = vi.spyOn(urlServices, "getURLStats");

    render(
      <StatsButton
        shortURL={shortURL}
        displayedStats={true}
        setDisplayedStats={setDisplayedStatsMock}
        setStats={setStatsMock}
      />,
    );

    const button = screen.getByRole("button", { name: "Shrink" });
    expect(button).toBeInTheDocument();

    fireEvent.click(button);

    expect(getStatsSpy).not.toHaveBeenCalled();
    expect(setDisplayedStatsMock).toHaveBeenCalled();
  });
});
