import type {
  CreateRemotePathMappingInput,
  RemotePathMapping,
  UpdateRemotePathMappingInput,
} from "../schemas"
import { fetchAPI } from "./client"

export const remotePathMappingApi = {
  list: () => fetchAPI<RemotePathMapping[]>("/remotepathmapping"),

  create: (data: CreateRemotePathMappingInput) =>
    fetchAPI<RemotePathMapping>("/remotepathmapping", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: UpdateRemotePathMappingInput) =>
    fetchAPI<RemotePathMapping>(`/remotepathmapping/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetchAPI<void>(`/remotepathmapping/${id}`, {
      method: "DELETE",
    }),
}
