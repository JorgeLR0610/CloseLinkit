import type { GetURLStatsAPIResponse, URLStats, ShortenURLAPIResponse } from "../types/url"
import request from "./apiClient"

export async function shortenURL(originalURL: string) {
    const data = await request<ShortenURLAPIResponse>(
        '/api/v1/shorten', {
        method: 'POST',
        body: JSON.stringify({
            url: originalURL
        })
    })

    return {
        shortURL: data.short_url
    }
}

export async function getURLStats(shortURL: string): Promise<URLStats> {
    const shortCode = shortURL.split('/').pop()

    const data = await request<GetURLStatsAPIResponse>(
        `/api/v1/${shortCode}/stats`, {
        method: 'GET'
    })

    return {
        originalURL: data.original_url,
        clickCount: data.click_count,
        createdAt: new Date(data.created_at)
    }
}