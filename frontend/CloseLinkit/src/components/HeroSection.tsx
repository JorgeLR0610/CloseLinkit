import { useState } from 'react'
import type { SyntheticEvent } from 'react'
import './HeroSection.css'

interface Props {
  onShorten: (url: string) => void;
}

export default function HeroSection({ onShorten }: Props) {
  const [url, setUrl] = useState('')

  const handleSubmit = (e: SyntheticEvent) => {
    e.preventDefault()
    onShorten(url)
    setUrl('')
  }

  return (
    <section className="hero fade-in">
      <h1 className="hero-title">URL shortening service</h1>
      <p className="hero-subtitle">Fast, secure, and incredibly easy to use.</p>
      
      <form className="hero-form glass-panel" onSubmit={handleSubmit}>
        <input 
          type="url" 
          required
          placeholder="Enter the link here" 
          className="hero-input"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
        <button type="submit" className="hero-btn">Shorten URL</button>
      </form>
    </section>
  )
}
