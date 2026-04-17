import * as z from "zod"

export const CustomProviderSettingSchema = z.object({
  name: z.string(),
  domain: z.string(),
  formatHint: z.string(),
})

export const GeneralSettingsSchema = z.object({
  apiKey: z.string(),
  bindAddress: z.string(),
  port: z.number(),
  dataDir: z.string(),
  libraryDir: z.string(),
  logLevel: z.string(),
  customProvidersEnabled: z.boolean(),
  customProvidersActive: z.array(CustomProviderSettingSchema),
})

export const RootFolderSchema = z.object({
  id: z.number(),
  path: z.string(),
  name: z.string(),
  defaultQualityProfileId: z.number().optional(),
  freeSpace: z.number().optional(),
  totalSpace: z.number().optional(),
  authorCount: z.number().optional(),
  accessible: z.boolean(),
})

export const CreateRootFolderInputSchema = z.object({
  path: z.string(),
  name: z.string(),
  defaultQualityProfileId: z.number().optional(),
})

export const UpdateRootFolderInputSchema = z.object({
  path: z.string().optional(),
  name: z.string().optional(),
  defaultQualityProfileId: z.number().optional(),
  moveFiles: z.boolean().optional(),
})

export const RemotePathMappingSchema = z.object({
  id: z.number(),
  host: z.string(),
  remotePath: z.string(),
  localPath: z.string(),
  createdAt: z.string(),
})

export const CreateRemotePathMappingInputSchema = z.object({
  host: z.string().optional(),
  remotePath: z.string(),
  localPath: z.string(),
})

export const UpdateRemotePathMappingInputSchema = z.object({
  host: z.string().optional(),
  remotePath: z.string().optional(),
  localPath: z.string().optional(),
})

export type CustomProviderSetting = z.infer<typeof CustomProviderSettingSchema>
export type GeneralSettings = z.infer<typeof GeneralSettingsSchema>
export type RootFolder = z.infer<typeof RootFolderSchema>
export type CreateRootFolderInput = z.infer<typeof CreateRootFolderInputSchema>
export type UpdateRootFolderInput = z.infer<typeof UpdateRootFolderInputSchema>
export type RemotePathMapping = z.infer<typeof RemotePathMappingSchema>
export type CreateRemotePathMappingInput = z.infer<typeof CreateRemotePathMappingInputSchema>
export type UpdateRemotePathMappingInput = z.infer<typeof UpdateRemotePathMappingInputSchema>
