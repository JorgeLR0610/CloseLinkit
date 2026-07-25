export type URLItem = {
    original: string
    short: string
}

// These fields must be set according to the API DTO, whose previous set fields
// should be changed
export type ShortenURLResponse = {
    shortURL: string
}