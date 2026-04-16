import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { wishlistApi } from "../lib/api/wishlist"

export function useWishlist() {
  return useQuery({
    queryKey: ["wishlist"],
    queryFn: wishlistApi.list,
    staleTime: 30 * 1000,
  })
}

export function useAddToWishlist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wishlistApi.add,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wishlist"] })
    },
  })
}

export function useRemoveFromWishlist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wishlistApi.remove,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wishlist"] })
    },
  })
}
