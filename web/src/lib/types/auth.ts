export interface User {
  id: number
  username: string
  role: string
  apiKey?: string
  createdAt: string
}

export interface LoginResponse {
  user: User
  apiKey: string
}
