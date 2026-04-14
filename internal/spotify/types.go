package spotify

type Image struct {
    URL    string `json:"url"`
    Height int    `json:"height"`
    Width  int    `json:"width"`
}

type Artist struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type Album struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    ReleaseDate string  `json:"release_date"`
    TotalTracks int     `json:"total_tracks"`
    Images      []Image `json:"images"`
    Artists     []Artist `json:"artists"`
}

type Track struct {
    ID         string   `json:"id"`
    Name       string   `json:"name"`
    Artists    []Artist `json:"artists"`
    Album      Album    `json:"album"`
    PreviewURL string   `json:"preview_url"`
    TrackNumber int     `json:"track_number"`
    ExternalIDs struct {
        ISRC string `json:"isrc"`
    } `json:"external_ids"`
}

type AlbumTracksPage struct {
    Items []Track `json:"items"`
    Total int     `json:"total"`
    Limit int     `json:"limit"`
    Offset int    `json:"offset"`
}

type PlaylistTrackItem struct {
    Track Track `json:"track"`
}

type PlaylistTracksPage struct {
    Items  []PlaylistTrackItem `json:"items"`
    Total  int                 `json:"total"`
    Limit  int                 `json:"limit"`
    Offset int                 `json:"offset"`
}
