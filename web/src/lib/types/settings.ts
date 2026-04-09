export interface GeneralSettings {
  apiKey: string
  bindAddress: string
  port: number
  dataDir: string
  libraryDir: string
  logLevel: string
}

export interface RootFolder {
  id: number
  path: string
  name: string
  defaultQualityProfileId?: number
  freeSpace?: number
  totalSpace?: number
  authorCount?: number
  accessible: boolean
}

export interface CreateRootFolderInput {
  path: string
  name: string
  defaultQualityProfileId?: number
}

export interface UpdateRootFolderInput {
  path?: string
  name?: string
  defaultQualityProfileId?: number
  moveFiles?: boolean
}
