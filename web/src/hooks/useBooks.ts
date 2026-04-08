import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { bookApi, type CreateBookInput, type ListBooksParams } from "../lib/api"

export function useBooks(params?: ListBooksParams) {
  return useQuery({
    queryKey: ["books", params],
    queryFn: () => bookApi.list(params),
    staleTime: 30 * 1000,
  })
}

export function useBook(id: number) {
  return useQuery({
    queryKey: ["book", id],
    queryFn: () => bookApi.get(id),
    enabled: id > 0,
    staleTime: 30 * 1000,
  })
}

export function useCreateBook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBookInput) => bookApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    },
  })
}

export function useUpdateBook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<CreateBookInput> }) =>
      bookApi.update(id, data),
    onSuccess: (book) => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
      queryClient.setQueryData(["book", book.id], book)
    },
  })
}

export function useDeleteBook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => bookApi.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
      queryClient.removeQueries({ queryKey: ["book", id] })
    },
  })
}
