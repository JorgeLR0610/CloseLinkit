import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import URLList from "./URLList";

describe("URLList Component", () => {
  it("renders list title and all items in history", () => {
    const history = [
      { originalURL: "https://example.com/one", shortURL: "http://localhost:8080/aaa" },
      { originalURL: "https://example.com/two", shortURL: "http://localhost:8080/bbb" },
    ];

    render(<URLList history={history} />);

    expect(
      screen.getByRole("heading", { level: 3, name: /list of shortened urls/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("https://example.com/one")).toBeInTheDocument();
    expect(screen.getByText("https://example.com/two")).toBeInTheDocument();
  });
});
