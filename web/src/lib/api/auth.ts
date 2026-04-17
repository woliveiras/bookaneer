import { LoginResponseSchema, UserSchema } from "../schemas/auth.schema"
import { fetchAPI } from "./client"

export const authApi = {
  login: (apiKey: string) =>
    fetchAPI(
      "/auth/me",
      { headers: { "X-Api-Key": apiKey } },
      UserSchema,
    ),

  loginWithCredentials: (username: string, password: string) =>
    fetchAPI(
      "/auth/login",
      { method: "POST", body: JSON.stringify({ username, password }) },
      LoginResponseSchema,
    ),

  me: () => fetchAPI("/auth/me", undefined, UserSchema),

  logout: () =>
    fetchAPI<{ status: string }>("/auth/logout", {
      method: "POST",
    }),
}
