import { fetchAPI } from "./client"
import type { User, LoginResponse } from "../types/auth"

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
