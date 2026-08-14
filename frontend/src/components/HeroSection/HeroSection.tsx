import { useState } from "react";
import type { SyntheticEvent } from "react";
import "./HeroSection.css";
import toast from "react-hot-toast";
import * as ipaddr from "ipaddr.js";

interface Props {
  onShorten: (url: string) => Promise<boolean>;
}

export default function HeroSection({ onShorten }: Props) {
  const [url, setURL] = useState("");
  const [error, setError] = useState<string | null>(null);

  function isValidHost(hostname: string): boolean {
    if (
      hostname === "localhost" ||
      hostname === "" ||
      !hostname.includes(".") ||
      hostname.endsWith(".")
    ) {
      return false;
    }

    try {
      const ip = ipaddr.parse(hostname);
      if (ip.range() === "loopback" || ip.range() === "linkLocal" || ip.range() === "private") {
        return false;
      }
    } catch {
      return true;
    }

    return true;
  }

  function isValidURL(url: string): boolean {
    const trimmedURL = url.trim();

    try {
      const parsedURL = new URL(trimmedURL);

      return (
        (parsedURL.protocol === "https:" || parsedURL.protocol === "http:") &&
        isValidHost(parsedURL.hostname)
      );
    } catch {
      return false;
    }
  }

  const handleSubmit = async (e: SyntheticEvent) => {
    e.preventDefault();

    if (!isValidURL(url)) {
      setError("The provided URL is invalid or malformed");
      return;
    }

    setError(null);

    const isSuccess = await onShorten(url);

    if (isSuccess) {
      setURL("");
      toast.success("Short URL created!");
    }
  };

  return (
    <section className="hero fade-in">
      <h1 className="hero-title">URL shortening service</h1>
      <p className="hero-subtitle">Fast, secure, and incredibly easy to use.</p>

      <form className="hero-form glass-panel" onSubmit={handleSubmit}>
        <input
          type="text"
          required
          placeholder="Paste your link here"
          className="hero-input"
          value={url}
          onChange={(e) => {
            setURL(e.target.value);
            if (error) setError(null);
          }}
        />

        <button type="submit" className="hero-btn">
          Shorten URL
        </button>
      </form>
      {error && (
        <p className="hero-error fade-in" role="alert">
          {error}
        </p>
      )}
    </section>
  );
}
