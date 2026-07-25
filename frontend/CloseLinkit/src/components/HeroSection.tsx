import { useState } from 'react'
import type { SyntheticEvent } from 'react'
import './HeroSection.css'

interface Props {
  onShorten: (url: string) => void;
}

export default function HeroSection({ onShorten }: Props) {
  const [url, setURL] = useState('')
  const [error, setError] = useState<string | null>(null)

  function isValidHttpURL(value: string): boolean {
    const trimmed = value.trim();

    if (!trimmed.startsWith("https://") && !trimmed.startsWith("http://")) {
      return false;
    }

    try {
      const url = new URL(trimmed);

      return (
        (url.protocol === "https:" || url.protocol === "http:") &&
        url.hostname.length > 0
      );
    } catch {
      return false;
    }
  }

  const handleSubmit = (e: SyntheticEvent) => {
    e.preventDefault()

    if (!isValidHttpURL(url)) {
      setError("The provided URL has a wrong format")
      return;
    }

    setError(null);
    onShorten(url);
    setURL('')
  }

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
            setURL(e.target.value)
            if (error) setError(null)
        }}
        />

        <button type="submit" className="hero-btn">
          Shorten URL
        </button>
      </form>
      {error && <p className="hero-error fade-in">{error}</p>}
    </section>
  )
}
