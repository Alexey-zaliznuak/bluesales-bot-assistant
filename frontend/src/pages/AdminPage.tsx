import { useMemo, useState } from 'react'
import { Chart } from '@gravity-ui/charts'
import { RangeDatePicker } from '@gravity-ui/date-components'
import { dateTime } from '@gravity-ui/date-utils'
import { Alert, Spin, Table, type TableColumnConfig } from '@gravity-ui/uikit'
import { useQuery } from '@tanstack/react-query'

import { api } from '../api/client'
import type { AdminUserUsage } from '../api/types'

const MOSCOW_TIMEZONE = 'Europe/Moscow'
const DATE_FORMAT = 'YYYY-MM-DD'

const numberFormatter = new Intl.NumberFormat('ru-RU')
const dateFormatter = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  timeZone: MOSCOW_TIMEZONE,
})

const userColumns: TableColumnConfig<AdminUserUsage>[] = [
  {
    id: 'login',
    name: 'Пользователь',
    primary: true,
    template: (item) => <span className="font-medium text-slate-100">{item.login}</span>,
  },
  {
    id: 'createdAt',
    name: 'Дата регистрации',
    template: (item) => dateFormatter.format(new Date(item.createdAt)),
  },
  {
    id: 'monthTokens',
    name: 'За текущий месяц',
    align: 'right',
    template: (item) => numberFormatter.format(item.monthTokens),
  },
  {
    id: 'allTimeTokens',
    name: 'За всё время',
    align: 'right',
    template: (item) => numberFormatter.format(item.allTimeTokens),
  },
]

export default function AdminPage() {
  const today = useMemo(() => dateTime({ timeZone: MOSCOW_TIMEZONE }).startOf('day'), [])
  const [range, setRange] = useState({
    start: today.subtract(29, 'day'),
    end: today,
  })

  const from = range.start.format(DATE_FORMAT)
  const to = range.end.format(DATE_FORMAT)
  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-dashboard', from, to],
    queryFn: () => api.adminDashboard(from, to),
  })

  const chartData = useMemo(
    () => ({
      series: {
        data: [
          {
            type: 'bar-x' as const,
            name: 'Токены',
            data: (data?.daily ?? []).map((item) => ({
              x: item.date,
              y: item.totalTokens,
            })),
          },
        ],
      },
    }),
    [data?.daily],
  )

  return (
    <div className="h-full overflow-y-auto bg-surface-950">
      <div className="mx-auto max-w-7xl space-y-6 px-6 py-7">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-slate-100">
              Использование системы
            </h1>
            <p className="mt-1 text-sm text-slate-500">
              Статистика токенов OpenRouter, часовой пояс — Москва
            </p>
          </div>
          <RangeDatePicker
            aria-label="Период статистики"
            value={range}
            onUpdate={(value) => value && setRange(value)}
            maxValue={today}
            timeZone={MOSCOW_TIMEZONE}
            format="DD.MM.YYYY"
            size="l"
          />
        </div>

        {error && (
          <Alert
            theme="danger"
            message={error instanceof Error ? error.message : 'Не удалось загрузить статистику'}
          />
        )}

        {isLoading ? (
          <div className="flex min-h-80 items-center justify-center">
            <Spin size="l" />
          </div>
        ) : (
          <>
            <section className="grid gap-4 sm:grid-cols-2">
              <MetricCard
                label="Токены за выбранный период"
                value={numberFormatter.format(data?.totalTokens ?? 0)}
                caption={`${formatShortDate(from)} — ${formatShortDate(to)}`}
              />
              <MetricCard
                label="Пользователей в системе"
                value={numberFormatter.format(data?.users.length ?? 0)}
                caption="Включая администратора"
              />
            </section>

            <section className="card p-6">
              <div className="mb-5">
                <h2 className="text-base font-semibold text-slate-100">Расход токенов по дням</h2>
                <p className="mt-1 text-sm text-slate-500">Суммарно по всем пользователям</p>
              </div>
              <div className="h-80">
                <Chart data={chartData} lang="ru" />
              </div>
            </section>

            <section className="card overflow-hidden">
              <div className="border-b border-surface-700 px-6 py-5">
                <h2 className="text-base font-semibold text-slate-100">Пользователи</h2>
                <p className="mt-1 text-sm text-slate-500">
                  Расход за текущий московский месяц и за всё время
                </p>
              </div>
              <div className="overflow-x-auto">
                <Table<AdminUserUsage>
                  className="min-w-[720px]"
                  columns={userColumns}
                  data={data?.users ?? []}
                  getRowDescriptor={(item) => ({ id: item.id })}
                  emptyMessage="Пользователей пока нет"
                  width="max"
                  edgePadding
                />
              </div>
            </section>
          </>
        )}
      </div>
    </div>
  )
}

function MetricCard({ label, value, caption }: { label: string; value: string; caption: string }) {
  return (
    <div className="card p-5">
      <div className="text-sm font-medium text-slate-500">{label}</div>
      <div className="mt-2 text-3xl font-semibold tracking-tight text-slate-100">{value}</div>
      <div className="mt-1 text-xs text-slate-400">{caption}</div>
    </div>
  )
}

function formatShortDate(value: string) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    timeZone: MOSCOW_TIMEZONE,
  }).format(new Date(`${value}T00:00:00+03:00`))
}
