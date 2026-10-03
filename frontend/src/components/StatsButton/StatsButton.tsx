import type { Dispatch, SetStateAction } from "react";
import { getURLStats } from "../../services/urls";
import type { URLStats } from "../../types/url";
import toast from "react-hot-toast";

interface Props {
  shortURL: string;
  className?: string;
  displayedStats: boolean;
  setDisplayedStats: Dispatch<SetStateAction<boolean>>;
  setStats: Dispatch<SetStateAction<URLStats | null>>;
  stats?: URLStats | null;
  accessToken?: string;
}

export default function StatsButton({
  shortURL,
  className = "",
  displayedStats,
  setDisplayedStats,
  setStats,
  stats,
  accessToken,
}: Props) {
  const handleStatsDisplay = async () => {
    if (displayedStats) {
      setDisplayedStats(false);
      return;
    }

    try {
      setStats(await getURLStats(shortURL, accessToken));
    } catch {
      if (!stats) {
        toast.error("Could not load stats. Please try again later.");
        return;
      }
    }
    setDisplayedStats(true);
  };

  return (
    <button onClick={handleStatsDisplay} className={className}>
      {displayedStats ? "Shrink" : "Analytics"}
    </button>
  );
}
