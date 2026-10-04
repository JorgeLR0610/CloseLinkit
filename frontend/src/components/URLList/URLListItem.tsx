import { useState } from "react";
import { FaCalendarDays, FaHandPointer } from "react-icons/fa6";
import CopyButton from "../CopyButton/CopyButton";
import StatsButton from "../StatsButton/StatsButton";
import type { URLStats, URLItem } from "../../types/url";

interface Props {
  item: URLItem;
  onDelete?: (shortURL: string) => Promise<void> | void;
  isAuthenticated?: boolean;
  accessToken?: string;
}

export default function URLListItem({
  item,
  onDelete,
  isAuthenticated = false,
  accessToken,
}: Props) {
  const [displayedStats, setDisplayedStats] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [stats, setStats] = useState<URLStats | null>(() => {
    if (item.clickCount !== undefined && item.createdAt) {
      return {
        originalURL: item.originalURL,
        clickCount: item.clickCount,
        createdAt: new Date(item.createdAt),
      };
    }
    return null;
  });

  const formattedDate = stats?.createdAt
    ? new Date(stats.createdAt).toLocaleString(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      })
    : "";

  const handleDelete = async () => {
    if (!onDelete || isDeleting) return;
    setIsDeleting(true);
    try {
      await onDelete(item.shortURL);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="url-item glass-panel">
      <div className="url-item-main">
        <span className="original-url" title={item.originalURL}>
          {item.originalURL}
        </span>

        <div className="short-url-group">
          <span className="short-url">
            <a href={item.shortURL} target="_blank" rel="noopener noreferrer">
              {item.shortURL}
            </a>
          </span>

          <CopyButton textToCopy={item.shortURL} className="util-btns-small" />

          {isAuthenticated && (
            <StatsButton
              shortURL={item.shortURL}
              className={`util-btns-small ${displayedStats ? "active" : ""}`}
              displayedStats={displayedStats}
              setDisplayedStats={setDisplayedStats}
              setStats={setStats}
              stats={stats}
              accessToken={accessToken}
            />
          )}

          {onDelete && (
            <button
              type="button"
              onClick={handleDelete}
              disabled={isDeleting}
              className="util-btns-small btn-delete"
              aria-label={isDeleting ? "Deleting URL" : "Delete URL"}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </button>
          )}
        </div>
      </div>

      {isAuthenticated && displayedStats && stats && (
        <div className="url-item-stats fade-in">
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-icon">
                <FaHandPointer />
              </div>
              <div className="stat-info">
                <span className="stat-label">Total Clicks</span>
                <span className="stat-value">{stats.clickCount}</span>
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon">
                <FaCalendarDays />
              </div>
              <div className="stat-info">
                <span className="stat-label">Created On</span>
                <span className="stat-value">{formattedDate}</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
