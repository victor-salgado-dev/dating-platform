'use client';

import { apiFetch, ApiError, FavoritesResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { InteractionListSection, InteractionCardItem, InteractionTabConfig } from '@/components/InteractionListSection';

type Tab = 'sent' | 'received' | 'mutual';

export default function FavoritesSection() {
  const { dictionary } = useI18n();

  async function fetchTab(tab: Tab, page: number) {
    const data = await apiFetch<FavoritesResponse>(`/favorites/${tab}?page=${page}`);
    const items: InteractionCardItem[] = data.items.map((item) => ({
      profile_id: item.profile_id,
      display_name: item.display_name,
      age: item.age,
      gender: item.gender,
      country_code: item.country_code,
      region: item.region,
      has_photo: item.has_photo,
      photo_url: item.photo_url,
      liked: item.liked,
      favorited: item.favorited,
      received_like: item.received_like,
      received_favorite: item.received_favorite,
      interaction_at: item.favorited_at,
    }));
    return { items, page: data.page, total_pages: data.total_pages, total: data.total };
  }

  const tabs: InteractionTabConfig<Tab>[] = [
    {
      key: 'sent',
      label: dictionary.favorites.tabSent,
      emptyTitle: dictionary.favorites.emptySent,
      emptySubtitle: dictionary.favorites.emptySentSubtitle,
      fetchPage: (page) => fetchTab('sent', page),
    },
    {
      key: 'received',
      label: dictionary.favorites.tabReceived,
      emptyTitle: dictionary.favorites.emptyReceived,
      emptySubtitle: dictionary.favorites.emptyReceivedSubtitle,
      fetchPage: (page) => fetchTab('received', page),
    },
    {
      key: 'mutual',
      label: dictionary.favorites.tabMutual,
      emptyTitle: dictionary.favorites.emptyMutual,
      emptySubtitle: dictionary.favorites.emptyMutualSubtitle,
      fetchPage: (page) => fetchTab('mutual', page),
    },
  ];

  return (
    <InteractionListSection
      tabs={tabs}
      defaultTab="sent"
      loadingLabel={dictionary.favorites.loading}
      unauthorizedError={dictionary.favorites.errorUnauthorized}
      genericError={dictionary.favorites.loadError}
      isUnauthorized={(err) => err instanceof ApiError && err.status === 401}
      paginationInfo={(page, totalPages, total) =>
        dictionary.favorites.paginationInfo
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
