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
  return { title: dictionary.legal.privacy.title };
}

export default function PrivacyPage() {
  const dictionary = getDictionary(resolveLocale());
  const t = dictionary.legal.privacy;

  return (
    <main className={styles.main}>
      <p
        className={styles.placeholderNotice}
        dangerouslySetInnerHTML={{
          __html: `<strong>${dictionary.legal.notice.title}</strong> ${t.noticeBody}`,
        }}
      />

      <h1>{t.title}</h1>
      <p className={styles.meta}>{t.meta}</p>

      <h2>{t.s1Title}</h2>
      <ul>
        {t.s1Items.map((item, i) => (
          <li key={i} dangerouslySetInnerHTML={{ __html: item }} />
        ))}
      </ul>
      <p>{t.s1Closing}</p>

      <h2>{t.s2Title}</h2>
      <p>{t.s2Body}</p>

      <h2>{t.s3Title}</h2>
      <ul>
        {t.s3Items.map((item, i) => (
          <li key={i}>{item}</li>
        ))}
      </ul>
      <p>{t.s3Closing}</p>

      <h2>{t.s4Title}</h2>
      <p>{t.s4Body}</p>

      <h2>{t.s5Title}</h2>
      <p>{t.s5Body}</p>

      <h2>{t.s6Title}</h2>
      <p dangerouslySetInnerHTML={{ __html: t.s6Body }} />

      <h2>{t.s7Title}</h2>
      <p>{t.s7Body}</p>

      <h2>{t.s8Title}</h2>
      <p>{t.s8Body}</p>
    </main>
  );
}
