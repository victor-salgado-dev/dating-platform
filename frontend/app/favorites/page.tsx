'use client';

import { useI18n } from '@/lib/i18n/context';
import FavoritesSection from '@/components/FavoritesSection';
import styles from '@/components/ListSection.module.css';

export default function FavoritesPage() {
  const { dictionary } = useI18n();
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.favorites.title}</h1>
      <FavoritesSection />
    </main>
  );
}
