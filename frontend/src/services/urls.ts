import type { GetURLStatsResponse, ShortenURLResponse } from "../types/url"
import request from "./apiClient"

export async function shortenURL(originalURL: string): Promise<ShortenURLResponse> {
    const data = await request<{
        short_url: string
    }>('/api/v1/shorten', {
        method: 'POST',
        body: JSON.stringify({
            url: originalURL
        })
    })

    return {
        shortURL: data.short_url
    }
}

export async function getURLStats(shortURL: string): Promise<GetURLStatsResponse> {
    const shortCode = shortURL.split('/').pop()

    const data = await request<{
        original_url: string
        click_count: number
        created_at: Date
    }>(`/api/v1/${shortCode}/stats`, {
        method: 'GET'
    })

    return {
        originalURL: data.original_url,
        clickCount: data.click_count,
        createdAt: data.created_at
    }
}