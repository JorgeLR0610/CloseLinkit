import type { ShortenURLResponse } from "../types/url";

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

export async function shortenURL(originalURL: string): Promise<ShortenURLResponse> {
  try {
    const response = await fetch(`${BASE_URL}/api/v1/shorten`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: originalURL }),
    });

    if (!response.ok) {
      let errorMessage = "An error occurred while shortening the URL. Please try again later";

      try {
        const errorData = await response.json();

        if (errorData && errorData.error) {
          errorMessage = errorData.error;
        }
      } catch (parseError) {
        console.error("Error parsing backend JSON:", parseError);
      }

      throw new Error(errorMessage);
    }

    const data = await response.json();
    return {
      shortURL: data.short_url,
    };
  } catch (error) {
    // When network or CORS errors occur, fetch usually throws a TypeError exception
    if (error instanceof TypeError) {
      // MDN itself warns that navigator.onLine is unreliable and should be used to provide
      // hints when the user may seem offline
      if (!navigator.onLine) {
        throw new Error("You seem to be offline. Please check your internet connection.");
      }
      throw new Error("Unable to reach the service. Please try again later.");
    }
    throw error;
  }
}
