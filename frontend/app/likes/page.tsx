'use client';

import { useI18n } from '@/lib/i18n/context';
import LikesSection from '@/components/LikesSection';
import styles from '@/components/ListSection.module.css';

export default function LikesPage() {
  const { dictionary } = useI18n();
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.likes.title}</h1>
      <LikesSection />
    </main>
  );
}
