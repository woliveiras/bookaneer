import * as z from "zod"

export const UserSchema = z.object({
  id: z.number(),
  username: z.string(),
  role: z.string(),
  apiKey: z.string().optional(),
  createdAt: z.string(),
})

export const LoginResponseSchema = z.object({
  user: UserSchema,
  apiKey: z.string(),
})

export type User = z.infer<typeof UserSchema>
export type LoginResponse = z.infer<typeof LoginResponseSchema>
