export interface NamingSettings {
  enabled: boolean
  authorFolderFormat: string
  bookFileFormat: string
  replaceSpaces: boolean
  colonReplacement: string
}

export interface NamingSettingsInput {
  enabled?: boolean
  authorFolderFormat?: string
  bookFileFormat?: string
  replaceSpaces?: boolean
  colonReplacement?: string
}

export interface NamingPreviewInput {
  authorFolderFormat: string
  bookFileFormat: string
  replaceSpaces: boolean
  colonReplacement: string
}

export interface NamingPreview {
  authorFolder: string
  filename: string
  relativePath: string
  fullPath: string
}

export interface RenameResult {
  total: number
  renamed: number
  skipped: number
  errors?: string[]
  files?: RenamedFile[]
}

export interface RenamedFile {
  bookId: number
  oldPath: string
  newPath: string
}
