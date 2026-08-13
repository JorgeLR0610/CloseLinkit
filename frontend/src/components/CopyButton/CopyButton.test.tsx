import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import CopyButton from "../CopyButton/CopyButton";

describe("CopyButton Component", () => {
  const textToCopy = "http://localhost:8080/xyz1234";

  beforeEach(() => {
    vi.useFakeTimers();
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('renders initially with text "Copy" and is enabled', () => {
    render(<CopyButton textToCopy={textToCopy} />);
    const button = screen.getByRole("button", { name: "Copy" });
    expect(button).toBeInTheDocument();
    expect(button).not.toBeDisabled();
  });

  it('copies text to clipboard, changes label to "Copied!", disables button, and resets after 2 seconds', async () => {
    render(<CopyButton textToCopy={textToCopy} />);

    const button = screen.getByRole("button", { name: "Copy" });

    await act(async () => {
      fireEvent.click(button);
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(textToCopy);

    const copiedButton = screen.getByRole("button", { name: "Copied!" });
    expect(copiedButton).toBeInTheDocument();
    expect(copiedButton).toBeDisabled();

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    const resetButton = screen.getByRole("button", { name: "Copy" });
    expect(resetButton).toBeInTheDocument();
    expect(resetButton).not.toBeDisabled();
  });
});
