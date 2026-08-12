import { useState } from 'react'
import { FaCalendarDays, FaHandPointer } from 'react-icons/fa6'
import CopyButton from '../UtilButtons/CopyButton'
import StatsButton from '../UtilButtons/StatsButton'
import type { URLStats, URLItem } from '../../types/url'

interface Props {
    item: URLItem
}

export default function URLListItem({ item }: Props) {
    const [displayedStats, setDisplayedStats] = useState(false)
    const [stats, setStats] = useState<URLStats | null>(null)

    const formattedDate = stats?.createdAt
        ? new Date(stats.createdAt).toLocaleString(undefined, {
            dateStyle: 'medium',
            timeStyle: 'short',
        })
        : ''

    return (
        <div className="url-item glass-panel">
            <div className="url-item-main">
                <span
                    className="original-url"
                    title={item.originalURL}
                >
                    {item.originalURL}
                </span>

                <div className="short-url-group">
                    <span className="short-url">
                        <a
                            href={item.shortURL}
                            target='_blank'
                            rel='noopener noreferrer'
                        >
                            {item.shortURL}
                        </a>
                    </span>

                    <CopyButton
                        textToCopy={item.shortURL}
                        className="util-btns-small"
                    />

                    <StatsButton
                        shortURL={item.shortURL}
                        className={`util-btns-small ${displayedStats ? 'active' : ''}`}
                        displayedStats={displayedStats}
                        setDisplayedStats={setDisplayedStats}
                        setStats={setStats}
                    />
                </div>
            </div>

            {displayedStats && stats && (
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
    )
}