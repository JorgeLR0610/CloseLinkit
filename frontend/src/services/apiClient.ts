const baseURL = import.meta.env.VITE_API_BASE_URL;

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  try {
    const response = await fetch(`${baseURL}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      let errorMessage = "An unexpected error occurred. Please try again later.";

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

    return (await response.json()) as T;
  } catch (error) {
    // fetch() usually throws TypeError for network/CORS errors
    if (error instanceof TypeError) {
      if (!navigator.onLine) {
        throw new Error("You seem to be offline. Please check your internet connection.");
      }

      throw new Error("Unable to reach the service. Please try again later.");
    }

    throw error;
  }
}

export default request;
