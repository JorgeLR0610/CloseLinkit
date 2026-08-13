import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import toast from "react-hot-toast";
import HeroSection from "./HeroSection";

describe("HeroSection Component", () => {
  const onShortenMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(toast, "success").mockImplementation(() => "");
  });

  it("renders initial elements: title, subtitle, input, and submit button", () => {
    render(<HeroSection onShorten={onShortenMock} />);

    expect(
      screen.getByRole("heading", { level: 1, name: /url shortening service/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/fast, secure, and incredibly easy to use\./i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/paste your link here/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /shorten url/i })).toBeInTheDocument();
  });

  it.each(["http://example.com", "https://example.com"])(
    'accepts valid URL "%s", calls onShorten, triggers success toast, and clears input',
    async (validUrl) => {
      onShortenMock.mockResolvedValue(true);

      render(<HeroSection onShorten={onShortenMock} />);
      const input = screen.getByPlaceholderText(/paste your link here/i);
      const submitButton = screen.getByRole("button", { name: /shorten url/i });

      fireEvent.change(input, { target: { value: validUrl } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(onShortenMock).toHaveBeenCalledWith(validUrl);
        expect(toast.success).toHaveBeenCalledWith("Short URL created!");
        expect(input).toHaveValue("");
      });
    },
  );

  it.each(["hello", "example.com", "ftp://example.com", "https://", "http://"])(
    'rejects invalid URL "%s", displays alert, and clears alert upon typing',
    (invalidUrl) => {
      render(<HeroSection onShorten={onShortenMock} />);
      const input = screen.getByPlaceholderText(/paste your link here/i);
      const submitButton = screen.getByRole("button", { name: /shorten url/i });

      fireEvent.change(input, { target: { value: invalidUrl } });
      fireEvent.click(submitButton);

      expect(onShortenMock).not.toHaveBeenCalled();

      const alert = screen.getByRole("alert");
      expect(alert).toBeInTheDocument();
      expect(alert).toHaveTextContent(/the provided url is invalid or malformed/i);

      // Typing into input clears the error
      fireEvent.change(input, { target: { value: invalidUrl + "a" } });
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    },
  );

  it("keeps input value when API onShorten returns false", async () => {
    onShortenMock.mockResolvedValue(false);

    render(<HeroSection onShorten={onShortenMock} />);
    const input = screen.getByPlaceholderText(/paste your link here/i);
    const submitButton = screen.getByRole("button", { name: /shorten url/i });

    const testUrl = "https://example.com";
    fireEvent.change(input, { target: { value: testUrl } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(onShortenMock).toHaveBeenCalledWith(testUrl);
      expect(input).toHaveValue(testUrl);
      expect(toast.success).not.toHaveBeenCalled();
    });
  });
});
