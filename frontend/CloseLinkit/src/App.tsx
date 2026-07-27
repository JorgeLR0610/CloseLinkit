import { useState } from 'react'
import Header from './components/Header'
import HeroSection from './components/HeroSection'
import RecentURLBox from './components/RecentURLBox'
import URLList from './components/URLList'
import FooterCTA from './components/FooterCTA'
import './App.css'
import type { URLItem } from './types/url'
import { shortenURL } from './api/urls'

function App() {
  const [recentURL, setRecentURL] = useState<string | null>(null)
  const [urlHistory, setURLHistory] = useState<URLItem[]>([])
  const [error, setError] = useState<string | null>(null)

  const handleShortenURL = async (originalURL: string) => {
    if (!originalURL.trim()) return;
    
    try {
      setError(null)
      const data = await shortenURL(originalURL)
      setRecentURL(data.shortURL)
      setURLHistory(urlHistory => [{ originalURL: originalURL, shortURL: data.shortURL }, ...urlHistory])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unexpected error')
    }
  }

  return (
    <div className="app-container">
      <Header />
      <main className="main-content">
        <HeroSection onShorten={handleShortenURL} />

        {error && (
          <p className="hero-error fade-in" role="alert">
            {error}
          </p>
        )}
        
        {recentURL && <RecentURLBox shortURL={recentURL} />}
        
        {urlHistory.length > 0 && <URLList history={urlHistory} />}
      </main>
      <FooterCTA />
    </div>
  )
}

export default App
