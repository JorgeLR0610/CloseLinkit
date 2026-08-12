import { useEffect, useState } from 'react'
import Header from './components/Header/Header'
import HeroSection from './components/HeroSection/HeroSection'
import RecentURLBox from './components/RecentURLBox/RecentURLBox'
import URLList from './components/URLList/URLList'
import './App.css'
import type { URLItem } from './types/url'
import { shortenURL } from './services/urls'
import toast from 'react-hot-toast'

function App() {
  const [urlHistory, setURLHistory] = useState<URLItem[]>(() => {
    try {
      const stored = localStorage.getItem("history")
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  })
  const [recentURL, setRecentURL] = useState<string | null>(null)

  useEffect(() => {
    localStorage.setItem(
      "history",
      JSON.stringify(urlHistory)
    )
  }, [urlHistory])

  const handleShortenURL = async (originalURL: string): Promise<boolean> => {
    if (!originalURL.trim()) return false;

    try {
      const data = await shortenURL(originalURL)
      setRecentURL(data.shortURL)

      // Save up to 10 short URLs in localStorage
      setURLHistory(previousHistory =>
        [
          {
            originalURL,
            shortURL: data.shortURL,
          },
          ...previousHistory,
        ].slice(0, 10)
      )

      return true;
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Unexpected error")
      return false;
    }
  }

  return (
    <div className="app-container">
      <Header />
      <main className="main-content">
        <HeroSection onShorten={handleShortenURL} />
        {recentURL && <RecentURLBox shortURL={recentURL} />}
        {urlHistory.length > 0 && <URLList history={urlHistory} />}
      </main>
    </div>
  )
}

export default App
