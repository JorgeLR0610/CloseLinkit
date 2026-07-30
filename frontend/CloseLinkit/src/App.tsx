import { useState } from 'react'
import Header from './components/Header/Header'
import HeroSection from './components/HeroSection/HeroSection'
import RecentURLBox from './components/RecentURLBox/RecentURLBox'
import URLList from './components/URLList/URLList'
import './App.css'
import type { URLItem } from './types/url'
import { shortenURL } from './services/urls'

function App() {
  const [recentURL, setRecentURL] = useState<string | null>(null)
  const [urlHistory, setURLHistory] = useState<URLItem[]>([])
  const [error, setError] = useState<string | null>(null)

  const handleShortenURL = async (originalURL: string): Promise<boolean> => {
    if (!originalURL.trim()) return false;
    
    try {
      setError(null)
      const data = await shortenURL(originalURL)
      setRecentURL(data.shortURL)
      setURLHistory(urlHistory => [{ originalURL: originalURL, shortURL: data.shortURL }, ...urlHistory])
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unexpected error')
      return false;
    }
  }

  return (
    <div className="app-container">
      <Header />
      <main className="main-content">
        <HeroSection 
          onShorten={handleShortenURL}
          onClearGlobalError={() => {
            if (error) setError(null)
          }}
        />

        {error && (
          <p className="hero-error fade-in" role="alert">
            {error}
          </p>
        )}
        
        {recentURL && <RecentURLBox shortURL={recentURL} />}
        
        {urlHistory.length > 0 && <URLList history={urlHistory} />}
      </main>
    </div>
  )
}

export default App
