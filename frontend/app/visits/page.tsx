'use client';

import { useI18n } from '@/lib/i18n/context';
import VisitsSection from '@/components/VisitsSection';
import styles from '@/components/ListSection.module.css';

export default function VisitsPage() {
  const { dictionary } = useI18n();
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.visits.title}</h1>
      <VisitsSection />
    </main>
  );
}
