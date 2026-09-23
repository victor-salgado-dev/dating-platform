'use client';

import { useRouter } from 'next/navigation';

import { useI18n } from './context';
import { LOCALES, LOCALE_COOKIE_NAME, type Locale } from './config';

const ONE_YEAR_SECONDS = 60 * 60 * 24 * 365;

export default function LanguageSwitcher() {
  const router = useRouter();
  const { locale, dictionary } = useI18n();

  function handleChange(next: Locale) {
    if (next === locale) return;
    document.cookie = `${LOCALE_COOKIE_NAME}=${next}; Path=/; Max-Age=${ONE_YEAR_SECONDS}; SameSite=Lax`;
    router.refresh();
  }

  return (
    <div
      role="group"
      aria-label={dictionary.languageSwitcher.label}
      style={{ display: 'inline-flex', gap: '0.25rem' }}
    >
      {LOCALES.map((code) => {
        const isActive = code === locale;
        return (
          <button
            key={code}
            type="button"
            onClick={() => handleChange(code)}
            aria-pressed={isActive}
            style={{
              border: '1px solid currentColor',
              background: 'transparent',
              padding: '2px 8px',
              borderRadius: 6,
              fontSize: '0.72rem',
              fontWeight: 700,
              letterSpacing: '0.04em',
              textTransform: 'uppercase',
              cursor: 'pointer',
              opacity: isActive ? 1 : 0.55,
            }}
          >
            {code}
          </button>
        );
      })}
    </div>
  );
}
