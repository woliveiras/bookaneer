import type {
  NamingPreview,
  NamingPreviewInput,
  NamingSettings,
  NamingSettingsInput,
  RenameResult,
} from "../types"
import { fetchAPI } from "./client"

export const namingApi = {
  getSettings: () => fetchAPI<NamingSettings>("/naming"),

  updateSettings: (data: NamingSettingsInput) =>
    fetchAPI<NamingSettings>("/naming", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  preview: (data: NamingPreviewInput) =>
    fetchAPI<NamingPreview>("/naming/preview", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  previewRenameAll: () =>
    fetchAPI<RenameResult>("/naming/rename/preview", {
      method: "POST",
    }),

  renameAll: () =>
    fetchAPI<RenameResult>("/naming/rename", {
      method: "POST",
    }),
}
