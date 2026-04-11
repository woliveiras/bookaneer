import { useState } from "react"
import { ChevronDown, ChevronRight } from "lucide-react"
import { AuthLayout } from "../components/layout/AppLayout"
import { DownloadClientList } from "../containers/download/DownloadClientList"
import { IndexerList } from "../containers/indexers/IndexerList"
import { IndexerOptions } from "../containers/indexers/IndexerOptions"
import { NamingSettings } from "../containers/settings/NamingSettings"
import { RootFolderList } from "../containers/settings/RootFolderSettings"
import { SettingsGeneral } from "../containers/settings/SettingsGeneral"

export function SettingsPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Settings</h2>
      <div className="space-y-4">
        <GeneralSettings />
        <RootFolderSettings />
        <NamingSettingsSection />
        <IndexerSettings />
        <DownloadClientSettings />
      </div>
    </AuthLayout>
  )
}

function GeneralSettings() {
  const [isOpen, setIsOpen] = useState(true)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">General</span>
        <span className="text-muted-foreground">{isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <SettingsGeneral />
        </div>
      )}
    </div>
  )
}

function RootFolderSettings() {
  const [isOpen, setIsOpen] = useState(true)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Root Folders</span>
        <span className="text-muted-foreground">{isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <RootFolderList />
        </div>
      )}
    </div>
  )
}

function NamingSettingsSection() {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">File Naming</span>
        <span className="text-muted-foreground">{isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <NamingSettings />
        </div>
      )}
    </div>
  )
}

function IndexerSettings() {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Indexers</span>
        <span className="text-muted-foreground">{isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t space-y-6">
          <IndexerList />
          <hr className="border-border" />
          <IndexerOptions />
        </div>
      )}
    </div>
  )
}

function DownloadClientSettings() {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Download Clients</span>
        <span className="text-muted-foreground">{isOpen ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <DownloadClientList />
        </div>
      )}
    </div>
  )
}
