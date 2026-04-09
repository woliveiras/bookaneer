export type IndexerType = "newznab" | "torznab"

// Preset configurations for popular ebook indexers
export interface IndexerPreset {
  id: string
  name: string
  type: IndexerType
  baseUrl: string
  apiPath: string
  categories: string
  description: string
}

export const INDEXER_PRESETS: { usenet: IndexerPreset[]; torrents: IndexerPreset[] } = {
  usenet: [
    {
      id: "nzbgeek",
      name: "NZBgeek",
      type: "newznab",
      baseUrl: "https://api.nzbgeek.info",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Popular Usenet indexer with ebooks",
    },
    {
      id: "drunkenslug",
      name: "DrunkenSlug",
      type: "newznab",
      baseUrl: "https://api.drunkenslug.com",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Usenet indexer with good ebook coverage",
    },
    {
      id: "nzbfinder",
      name: "NZBFinder",
      type: "newznab",
      baseUrl: "https://nzbfinder.ws",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Dutch Usenet indexer with ebooks",
    },
    {
      id: "newznab-custom",
      name: "Newznab",
      type: "newznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "7000",
      description: "Custom Newznab-compatible indexer",
    },
  ],
  torrents: [
    {
      id: "myanonamouse",
      name: "MyAnonamouse",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "8000,8010",
      description: "Private tracker for ebooks (via Prowlarr/Jackett)",
    },
    {
      id: "bibliotik",
      name: "BiblioTik",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "8000,8010",
      description: "Private ebook tracker (via Prowlarr/Jackett)",
    },
    {
      id: "prowlarr",
      name: "Prowlarr",
      type: "torznab",
      baseUrl: "http://localhost:9696",
      apiPath: "/1/api",
      categories: "",
      description: "Indexer manager/proxy for Servarr",
    },
    {
      id: "jackett",
      name: "Jackett",
      type: "torznab",
      baseUrl: "http://localhost:9117",
      apiPath: "/api/v2.0/indexers/all/results/torznab",
      categories: "",
      description: "Torznab proxy for many trackers",
    },
    {
      id: "torznab-custom",
      name: "Torznab",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "",
      description: "Custom Torznab-compatible indexer",
    },
  ],
}
