import { AuthLayout } from "../components/layout/AppLayout"
import { BlocklistList } from "../containers/wanted/BlocklistList"
import { HistoryList } from "../containers/wanted/HistoryList"
import { QueueList } from "../containers/wanted/QueueList"

interface ActivityPageProps {
  tab: string
  onTabChange: (tab: string) => void
}

export function ActivityPage({ tab, onTabChange }: ActivityPageProps) {
  const tabs = [
    { id: "queue", label: "Queue" },
    { id: "history", label: "History" },
    { id: "blocklist", label: "Blocklist" },
  ]

  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Activity</h2>

      {/* Tab navigation */}
      <div className="border-b border-border mb-6">
        <nav className="-mb-px flex gap-1 overflow-x-auto" aria-label="Activity tabs">
          {tabs.map((t) => (
            <button
              type="button"
              key={t.id}
              onClick={() => onTabChange(t.id)}
              className={`
                whitespace-nowrap border-b-2 py-3 px-4 text-sm font-medium transition-colors min-h-[44px]
                ${
                  tab === t.id
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:border-muted-foreground/30 hover:text-foreground"
                }
              `}
              aria-current={tab === t.id ? "page" : undefined}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      {tab === "queue" && <QueueList />}
      {tab === "history" && <HistoryList />}
      {tab === "blocklist" && <BlocklistList />}
    </AuthLayout>
  )
}
