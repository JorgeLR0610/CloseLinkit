import type { GetURLStatsResponse, ShortenURLResponse } from "../types/url"

const baseURL = import.meta.env.VITE_API_BASE_URL

export async function shortenURL(originalURL: string): Promise<ShortenURLResponse> {
    try {
        const response = await fetch(`${baseURL}/api/v1/shorten`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: originalURL })
        })

        if (!response.ok) {
            let errorMessage = 'An error occurred while shortening the URL. Please try again later';

            try {
                const errorData = await response.json();

                if (errorData && errorData.error) {
                    errorMessage = errorData.error;
                }
            } catch (parseError) {
                console.error('Error parsing backend JSON:', parseError)
            }

            throw new Error(errorMessage)
        }

        const data = await response.json()
        return {
            shortURL: data.short_url
        }

    } catch (error) {
        // When network or CORS errors occur, fetch usually throws a TypeError exception
        if (error instanceof TypeError) {
            // MDN itself warns that navigator.onLine is unreliable and should be used to provide
            // hints when the user may seem offline
            if (!navigator.onLine) {
                throw new Error('You seem to be offline. Please check your internet connection.');
            }
            throw new Error('Unable to reach the service. Please try again later.');
        }
        throw error;
    }
}

// Consider extracting the request logic into another function (for both shortenURL and this function)
export async function getURLStats(shortURL: string): Promise<GetURLStatsResponse> {
    try {
        const shortCode = shortURL.split("/").pop()
        const fullURL = `${baseURL}/api/v1/${shortCode}/stats`

        const response = await fetch(`${fullURL}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        })

        if (!response.ok) {
            let errorMessage = 'An error occurred while retrieving the URL stats. Please try again later';

            try {
                const errorData = await response.json();

                if (errorData && errorData.error) {
                    errorMessage = errorData.error;
                }
            } catch (parseError) {
                console.error('Error parsing backend JSON:', parseError)
            }

            throw new Error(errorMessage)
        }

        const data = await response.json()
        return {
            originalURL: data.original_url,
            clickCount: data.click_count,
            createdAt: data.created_at
        }

    } catch (error) {
        // When network or CORS errors occur, fetch usually throws a TypeError exception
        if (error instanceof TypeError) {
            // MDN itself warns that navigator.onLine is unreliable and should be used to provide
            // hints when the user may seem offline
            if (!navigator.onLine) {
                throw new Error('You seem to be offline. Please check your internet connection.');
            }
            throw new Error('Unable to reach the service. Please try again later.');
        }
        throw error;
    }
}