import type { LoginResponse, User } from "../types/auth"
import { fetchAPI } from "./client"

export const authApi = {
  login: (apiKey: string) =>
    fetchAPI<User>("/auth/me", {
      headers: { "X-Api-Key": apiKey },
    }),

  loginWithCredentials: (username: string, password: string) =>
    fetchAPI<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),

  me: () => fetchAPI<User>("/auth/me"),

  logout: () =>
    fetchAPI<{ status: string }>("/auth/logout", {
      method: "POST",
    }),
}
