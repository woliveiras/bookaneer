import { create } from "zustand"
import { devtools } from "zustand/middleware"
import { immer } from "zustand/middleware/immer"
import {
  DEFAULT_SETTINGS,
  loadSettings,
  type ReaderSettings,
  saveSettings,
} from "../../components/reader/readerConfig"

interface ReaderSettingsActions {
  updateSettings: (updates: Partial<ReaderSettings>) => void
  reset: () => void
}

export const useReaderSettingsStore = create<ReaderSettings & ReaderSettingsActions>()(
  devtools(
    immer((set) => ({
      ...loadSettings(),

      updateSettings: (updates) =>
        set((state) => {
          Object.assign(state, updates)
          saveSettings({ ...state, ...updates })
        }),

      reset: () =>
        set(() => {
          saveSettings(DEFAULT_SETTINGS)
          return { ...DEFAULT_SETTINGS }
        }),
    })),
    { name: "ReaderSettingsStore", enabled: import.meta.env.DEV },
  ),
)
