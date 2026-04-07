import { useQuery } from "@tanstack/react-query"
import { settingsApi } from "../lib/api"

export function useGeneralSettings() {
  return useQuery({
    queryKey: ["settings", "general"],
    queryFn: () => settingsApi.getGeneral(),
  })
}
