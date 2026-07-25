import { useState } from 'react'
import Header from './components/Header'
import HeroSection from './components/HeroSection'
import RecentURLBox from './components/RecentURLBox'
import URLList from './components/URLList'
import FooterCTA from './components/FooterCTA'
import './App.css'
import type { URLItem } from './types/url'

function App() {
  const [recentURL, setRecentURL] = useState<string | null>(null)
  const [urlHistory, setUrlHistory] = useState<URLItem[]>([])

  const handleShortenURL = (originalURL: string) => {
    if (!originalURL.trim()) return;
    
    // API mock
    const fakeShort = `https://closelink.it/${Math.random().toString(36).substring(2, 7)}`
    setRecentURL(fakeShort)
    setUrlHistory([{ original: originalURL, short: fakeShort }, ...urlHistory])
  }

  return (
    <div className="app-container">
      <Header />
      <main className="main-content">
        <HeroSection onShorten={handleShortenURL} />
        
        
        {recentURL && <RecentURLBox shortURL={recentURL} />}
        
        {urlHistory.length > 0 && <URLList history={urlHistory} />}
      </main>
      <FooterCTA />
    </div>
  )
}

export default App
