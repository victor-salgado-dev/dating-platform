import type { Metadata } from 'next';
import { cookies } from 'next/headers';

import { DEFAULT_LOCALE, LOCALE_COOKIE_NAME, isValidLocale, type Locale } from '@/lib/i18n/config';
import { getDictionary } from '@/lib/i18n/get-dictionary';
import styles from '../legal.module.css';

function resolveLocale(): Locale {
  const cookieLocale = cookies().get(LOCALE_COOKIE_NAME)?.value ?? '';
  return isValidLocale(cookieLocale) ? cookieLocale : DEFAULT_LOCALE;
}

export async function generateMetadata(): Promise<Metadata> {
  const dictionary = getDictionary(resolveLocale());
  return { title: dictionary.legal.impressum.title };
}

export default function ImpressumPage() {
  const dictionary = getDictionary(resolveLocale());
  const t = dictionary.legal.impressum;

  return (
    <main className={styles.main}>
      <p
        className={styles.placeholderNotice}
        dangerouslySetInnerHTML={{
          __html: `<strong>${dictionary.legal.notice.title}</strong> ${t.noticeBody}`,
        }}
      />

      <h1>{t.title}</h1>

      <h2>{t.ownerTitle}</h2>
      <p dangerouslySetInnerHTML={{ __html: t.ownerBody }} />

      <h2>{t.contactTitle}</h2>
      <p dangerouslySetInnerHTML={{ __html: t.contactBody }} />

      <h2>{t.registrationTitle}</h2>
      <p>{t.registrationBody}</p>

      <h2>{t.editorialTitle}</h2>
      <p>{t.editorialBody}</p>

      <h2>{t.disputeTitle}</h2>
      <p>{t.disputeBody}</p>
    </main>
  );
}
