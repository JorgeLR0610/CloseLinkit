export type URLItem = {
  originalURL: string;
  shortURL: string;
};

export type ShortenURLResponse = {
<<<<<<< HEAD
  shortURL: string;
};
=======
    shortURL: string
}

export type GetURLStatsResponse = {
    originalURL: string
    clickCount: number
    createdAt: Date
}
>>>>>>> b4c95eb (feat(web): implement basic URL statistics section and refactor components for better structure)
