export interface CustomProviderSetting {
  name: string
  domain: string
  formatHint: string
}

export interface GeneralSettings {
  apiKey: string
  bindAddress: string
  port: number
  dataDir: string
  libraryDir: string
  logLevel: string
  customProvidersEnabled: boolean
  customProvidersActive: CustomProviderSetting[]
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

export interface RemotePathMapping {
  id: number
  host: string
  remotePath: string
  localPath: string
  createdAt: string
}

export interface CreateRemotePathMappingInput {
  host?: string
  remotePath: string
  localPath: string
}

export interface UpdateRemotePathMappingInput {
  host?: string
  remotePath?: string
  localPath?: string
}
