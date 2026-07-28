import type { ShortenURLResponse } from "../types/url"

const BASE_URL = import.meta.env.VITE_API_BASE_URL

export async function shortenURL(originalURL: string): Promise<ShortenURLResponse> {
    const response = await fetch(`${BASE_URL}/api/v1/shorten`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url: originalURL })
    })

    if (!response.ok) {
        const message = await response.text()
        throw new Error(message || `Request failed with status ${response.status}`)
    }

    const data = await response.json()
    return {
        shortURL: data.short_url
    }
}