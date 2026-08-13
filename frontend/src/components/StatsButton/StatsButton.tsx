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
}

export default function StatsButton({
  shortURL,
  className = "",
  displayedStats,
  setDisplayedStats,
  setStats,
}: Props) {
  const handleStatsDisplay = async () => {
    if (!displayedStats) {
      try {
        const resp = await getURLStats(shortURL);
        setStats(resp);
      } catch {
        toast.error("Could not load stats. Please try again later.");
        return;
      }
    }

    setDisplayedStats((prev) => !prev);
  };

  return (
    <button onClick={handleStatsDisplay} className={className}>
      {displayedStats ? "Shrink" : "Analytics"}
    </button>
  );
}
