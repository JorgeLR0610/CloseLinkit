import { useState } from 'react'
import CopyButton from '../UtilButtons/CopyButton'
import StatsButton from '../UtilButtons/StatsButton'
import type { GetURLStatsResponse, URLItem } from '../../types/url'

interface Props {
    item: URLItem
}

export default function URLListItem({ item }: Props) {
    const [displayedStats, setDisplayedStats] = useState(false)
    const [stats, setStats] = useState<GetURLStatsResponse | null>(null)

    return (
        <div className="url-item glass-panel">
            <span
                className="original-url"
                title={item.originalURL}
            >
                {item.originalURL}
            </span>

            <div className="short-url-group">
                <span className="short-url">
                    {item.shortURL}
                </span>

                <CopyButton
                    textToCopy={item.shortURL}
                    className="util-btns-small"
                />

                <StatsButton
                    shortURL={item.shortURL}
                    className="util-btns-small"
                    displayedStats={displayedStats}
                    setDisplayedStats={setDisplayedStats}
                    setStats={setStats}
                />
            </div>

            {displayedStats && stats && (
                <div>
                    <div>
                        Clicks: {stats.clickCount}
                    </div>
                    <div>
                        Created on: {stats.createdAt.toString()}
                    </div>
                </div>

            )}
        </div>
    )
}