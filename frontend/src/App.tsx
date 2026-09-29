import { useEffect, useState } from "react";
import { Routes, Route, Navigate } from "react-router";
import Header from "./components/Header/Header";
import FooterCTA from "./components/Footer/FooterCTA";
import HeroSection from "./components/HeroSection/HeroSection";
import RecentURLBox from "./components/RecentURLBox/RecentURLBox";
import URLList from "./components/URLList/URLList";
import LoginPage from "./pages/LoginPage/LoginPage";
import SignupPage from "./pages/SignupPage/SignupPage";
import "./App.css";
import type { URLItem } from "./types/url";
import type { UserURLItem } from "./types/auth";
import { shortenURL } from "./services/urls";
import { getUserURLs } from "./services/auth";
import { useAuth } from "./context/useAuth";
import toast from "react-hot-toast";

export function HomePage() {
  const { accessToken, isAuthenticated } = useAuth();
  const [localHistory, setLocalHistory] = useState<URLItem[]>(() => {
    try {
      const stored = localStorage.getItem("history");
      if (!stored) return [];
      const parsed: URLItem[] = JSON.parse(stored);
      const now = Date.now();
      return parsed.filter((item) => !item.expiresAt || new Date(item.expiresAt).getTime() > now);
    } catch {
      return [];
    }
  });
  const [userURLs, setUserURLs] = useState<UserURLItem[]>([]);
  const [recentURL, setRecentURL] = useState<string | null>(null);

  const [prevIsAuthenticated, setPrevIsAuthenticated] = useState(isAuthenticated);
  if (prevIsAuthenticated !== isAuthenticated) {
    setPrevIsAuthenticated(isAuthenticated);
    if (isAuthenticated) {
      setLocalHistory([]);
    }
  }

  // Sync anonymous URLs to localStorage only for unauthenticated users
  useEffect(() => {
    if (!isAuthenticated) {
      localStorage.setItem("history", JSON.stringify(localHistory));
    }
  }, [localHistory, isAuthenticated]);

  // Load authenticated user URLs from API
  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      return;
    }

    let isMounted = true;
    getUserURLs(accessToken)
      .then((data) => {
        if (isMounted) setUserURLs(data);
      })
      .catch((err) => {
        if (isMounted) {
          toast.error(err instanceof Error ? err.message : "Failed to load URLs.");
        }
      });

    return () => {
      isMounted = false;
    };
  }, [isAuthenticated, accessToken]);

  const handleShortenURL = async (originalURL: string): Promise<boolean> => {
    if (!originalURL.trim()) return false;

    try {
      const data = await shortenURL(originalURL, accessToken || undefined);
      setRecentURL(data.shortURL);

      if (isAuthenticated && accessToken) {
        try {
          const updated = await getUserURLs(accessToken);
          setUserURLs(updated);
        } catch {
          // If re-fetch fails, prepend optimistic item
          setUserURLs((prev) => [
            {
              original_url: originalURL,
              short_code: data.shortURL.split("/").pop() || "",
              short_url: data.shortURL,
              created_at: new Date().toISOString(),
              click_count: 0,
            },
            ...prev,
          ]);
        }
      } else {
        // Save up to 10 short URLs in localStorage for guests
        setLocalHistory((previousHistory) =>
          [
            {
              originalURL,
              shortURL: data.shortURL,
              expiresAt: data.expiresAt,
            },
            ...previousHistory,
          ].slice(0, 10),
        );
      }

      return true;
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Unexpected error");
      return false;
    }
  };

  const displayHistory: URLItem[] = isAuthenticated
    ? userURLs.map((u) => ({
        originalURL: u.original_url,
        shortURL: u.short_url,
        expiresAt: null,
        clickCount: u.click_count,
        createdAt: u.created_at,
      }))
    : localHistory;

  return (
    <>
      <main className="main-content">
        <HeroSection onShorten={handleShortenURL} />
        {recentURL && <RecentURLBox shortURL={recentURL} />}
        {displayHistory.length > 0 && <URLList history={displayHistory} />}
      </main>
      <FooterCTA />
    </>
  );
}

function App() {
  return (
    <div className="app-container">
      <Header />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/dashboard" element={<Navigate to="/" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}

export default App;
