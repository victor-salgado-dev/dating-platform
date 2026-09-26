'use client';

import { apiFetch, ApiError, VisitsResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { InteractionListSection, InteractionCardItem, InteractionTabConfig } from '@/components/InteractionListSection';

type Tab = 'received' | 'sent' | 'mutual';

export default function VisitsSection() {
  const { dictionary } = useI18n();

  async function fetchTab(tab: Tab, page: number) {
    const data = await apiFetch<VisitsResponse>(`/visits/${tab}?page=${page}`);
    const items: InteractionCardItem[] = data.items.map((item) => ({
      profile_id: item.profile_id,
      display_name: item.display_name,
      age: item.age,
      gender: item.gender,
      country_code: item.country_code,
      region: item.region,
      has_photo: item.has_photo,
      photo_url: item.photo_url,
      interaction_at: item.visited_at,
    }));
    return { items, page: data.page, total_pages: data.total_pages, total: data.total };
  }

  const tabs: InteractionTabConfig<Tab>[] = [
    {
      key: 'received',
      label: dictionary.visits.tabReceived,
      emptyTitle: dictionary.visits.emptyReceived,
      emptySubtitle: dictionary.visits.emptyReceivedSubtitle,
      fetchPage: (page) => fetchTab('received', page),
    },
    {
      key: 'sent',
      label: dictionary.visits.tabSent,
      emptyTitle: dictionary.visits.emptySent,
      emptySubtitle: dictionary.visits.emptySentSubtitle,
      fetchPage: (page) => fetchTab('sent', page),
    },
    {
      key: 'mutual',
      label: dictionary.visits.tabMutual,
      emptyTitle: dictionary.visits.emptyMutual,
      emptySubtitle: dictionary.visits.emptyMutualSubtitle,
      fetchPage: (page) => fetchTab('mutual', page),
    },
  ];

  return (
    <InteractionListSection
      tabs={tabs}
      defaultTab="received"
      loadingLabel={dictionary.visits.loading}
      unauthorizedError={dictionary.visits.errorUnauthorized}
      genericError={dictionary.visits.loadError}
      isUnauthorized={(err) => err instanceof ApiError && err.status === 401}
      paginationInfo={(page, totalPages, total) =>
        dictionary.visits.paginationInfo
          .replace('{page}', String(page))
          .replace('{totalPages}', String(totalPages))
          .replace('{total}', String(total))
      }
      sortLabel={dictionary.common.sortLabel}
      sortNewestFirst={dictionary.common.sortNewestFirst}
      sortOldestFirst={dictionary.common.sortOldestFirst}
      paginationPrev={dictionary.common.paginationPrev}
      paginationNext={dictionary.common.paginationNext}
    />
  );
}
