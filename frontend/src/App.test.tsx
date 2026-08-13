import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import toast from "react-hot-toast";
import App from "./App";
import * as urlServices from "./services/urls";

describe("App Component", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.spyOn(toast, "success").mockImplementation(() => "");
    vi.spyOn(toast, "error").mockImplementation(() => "");
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("renders initial state without recent box or history list when localStorage is empty", () => {
    render(<App />);

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/paste your link here/i)).toBeInTheDocument();
    expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
  });

  it("loads and displays initial history from localStorage", () => {
    const initialHistory = [
      { originalURL: "https://example.com/stored1", shortURL: "http://localhost:8080/s1" },
      { originalURL: "https://example.com/stored2", shortURL: "http://localhost:8080/s2" },
    ];
    localStorage.setItem("history", JSON.stringify(initialHistory));

    render(<App />);

    expect(screen.getByText(/list of shortened urls/i)).toBeInTheDocument();
    expect(screen.getByText("https://example.com/stored1")).toBeInTheDocument();
    expect(screen.getByText("https://example.com/stored2")).toBeInTheDocument();
  });

  it("gracefully handles corrupt JSON in localStorage", () => {
    localStorage.setItem("history", "invalid-json{{{");

    render(<App />);

    expect(screen.getByText("CloseLinkit")).toBeInTheDocument();
    expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
  });

  it("shortens URL, displays recent box, puts new item on top of history, and saves to localStorage", async () => {
    vi.spyOn(urlServices, "shortenURL").mockResolvedValue({
      shortURL: "http://localhost:8080/new123",
    });

    render(<App />);

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/target" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      // Recent URL Box rendered
      expect(screen.getAllByText("http://localhost:8080/new123").length).toBeGreaterThanOrEqual(1);
      // URL List rendered with original and short URL
      expect(screen.getByText("https://example.com/target")).toBeInTheDocument();
      // Ensure useEffect has flushed to localStorage
      const stored = JSON.parse(localStorage.getItem("history") || "[]");
      expect(stored).toEqual([
        {
          originalURL: "https://example.com/target",
          shortURL: "http://localhost:8080/new123",
        },
      ]);
    });
  });

  it("maintains a maximum of 10 items in history", async () => {
    const existing10Items = Array.from({ length: 10 }, (_, i) => ({
      originalURL: `https://example.com/item-${i + 1}`,
      shortURL: `http://localhost:8080/code-${i + 1}`,
    }));
    localStorage.setItem("history", JSON.stringify(existing10Items));

    vi.spyOn(urlServices, "shortenURL").mockResolvedValue({
      shortURL: "http://localhost:8080/code-11",
    });

    render(<App />);

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/item-11" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText("https://example.com/item-11")).toBeInTheDocument();
      const stored = JSON.parse(localStorage.getItem("history") || "[]");
      expect(stored).toHaveLength(10);
      expect(stored[0]).toEqual({
        originalURL: "https://example.com/item-11",
        shortURL: "http://localhost:8080/code-11",
      });
      // 10th item is preserved, oldest (item-10) was dropped
      expect(stored[9]).toEqual({
        originalURL: "https://example.com/item-9",
        shortURL: "http://localhost:8080/code-9",
      });
    });
  });

  it("handles shortenURL failure with toast error without modifying state", async () => {
    vi.spyOn(urlServices, "shortenURL").mockRejectedValue(new Error("Network error on API"));

    render(<App />);

    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitBtn = screen.getByRole("button", { name: /shorten url/i });

    fireEvent.change(input, { target: { value: "https://example.com/failed" } });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Network error on API");
      expect(screen.queryByText(/list of shortened urls/i)).not.toBeInTheDocument();
    });
  });
});
