import * as z from "zod"

export const NamingSettingsSchema = z.object({
  enabled: z.boolean(),
  authorFolderFormat: z.string(),
  bookFileFormat: z.string(),
  replaceSpaces: z.boolean(),
  colonReplacement: z.string(),
})

export const NamingSettingsInputSchema = NamingSettingsSchema.partial()

export const NamingPreviewInputSchema = z.object({
  authorFolderFormat: z.string(),
  bookFileFormat: z.string(),
  replaceSpaces: z.boolean(),
  colonReplacement: z.string(),
})

export const NamingPreviewSchema = z.object({
  authorFolder: z.string(),
  filename: z.string(),
  relativePath: z.string(),
  fullPath: z.string(),
})

export const RenamedFileSchema = z.object({
  bookId: z.number(),
  oldPath: z.string(),
  newPath: z.string(),
})

export const RenameResultSchema = z.object({
  total: z.number(),
  renamed: z.number(),
  skipped: z.number(),
  errors: z.array(z.string()).optional(),
  files: z.array(RenamedFileSchema).optional(),
})

export type NamingSettings = z.infer<typeof NamingSettingsSchema>
export type NamingSettingsInput = z.infer<typeof NamingSettingsInputSchema>
export type NamingPreviewInput = z.infer<typeof NamingPreviewInputSchema>
export type NamingPreview = z.infer<typeof NamingPreviewSchema>
export type RenamedFile = z.infer<typeof RenamedFileSchema>
export type RenameResult = z.infer<typeof RenameResultSchema>
