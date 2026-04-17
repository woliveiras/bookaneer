import type {
  CreateRootFolderInput,
  GeneralSettings,
  RootFolder,
  UpdateRootFolderInput,
} from "../schemas"
import { API_BASE, fetchAPI, getStoredApiKey } from "./client"

export const settingsApi = {
  getGeneral: () => fetchAPI<GeneralSettings>("/settings/general"),
}

export const rootFolderApi = {
  list: () => fetchAPI<RootFolder[]>("/rootfolder"),

  get: (id: number) => fetchAPI<RootFolder>(`/rootfolder/${id}`),

  create: (data: CreateRootFolderInput) =>
    fetchAPI<RootFolder>("/rootfolder", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: UpdateRootFolderInput) =>
    fetchAPI<RootFolder>(`/rootfolder/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetch(`${API_BASE}/rootfolder/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete root folder")
    }),
}
