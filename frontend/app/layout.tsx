import type { Metadata } from 'next';
import { cookies } from 'next/headers';
import { Fraunces, Public_Sans } from 'next/font/google';
import Link from 'next/link';
import HeaderChrome, { type NavLinkItem, type FilterLinkItem } from './header-chrome';
import LanguageSwitcher from '@/lib/i18n/LanguageSwitcher';
import { DEFAULT_LOCALE, LOCALE_COOKIE_NAME, isValidLocale, type Locale } from '@/lib/i18n/config';
import { getDictionary } from '@/lib/i18n/get-dictionary';
import { I18nProvider } from '@/lib/i18n/context';
import './globals.css';
import styles from './layout.module.css';

// Serif cálida para marca/títulos + sans humanista muy legible para el resto.
// Se exponen como variables CSS (--font-voice / --font-ui) y se consumen
// desde globals.css y layout.module.css.
const fraunces = Fraunces({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-voice',
  display: 'swap',
});

const publicSans = Public_Sans({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-ui',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Dating Platform',
  description: 'Plataforma internacional de dating/relaciones',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = cookies();
  const cookieLocale = cookieStore.get(LOCALE_COOKIE_NAME)?.value ?? '';
  const locale: Locale = isValidLocale(cookieLocale) ? cookieLocale : DEFAULT_LOCALE;
  const dictionary = getDictionary(locale);

  // Likes y Visits ya no van en el menú principal: ahora viven como pestañas
  // dentro de Activity (junto con Favorites). Ver activity/page.tsx.
  const navLinks: NavLinkItem[] = [
    { href: '/', label: dictionary.nav.home },
    { href: '/discover', label: dictionary.nav.discover },
    { href: '/activity', label: dictionary.nav.activity },
    { href: '/matches', label: dictionary.nav.matches },
    { href: '/messages', label: dictionary.nav.messages },
  ];

  // Popular / Online now / New members ahora son pestañas sin recarga dentro
  // de la propia Home (page.tsx), no rutas distintas, así que ya no hace
  // falta esta barra global de filtros. Se deja el array vacío en vez de
  // borrar la prop, por si en el futuro se necesita para otra sección.
  const filterLinks: FilterLinkItem[] = [];

  return (
    <html lang={locale} className={`${fraunces.variable} ${publicSans.variable}`}>
      <body>
        <I18nProvider locale={locale} dictionary={dictionary}>
          <HeaderChrome navLinks={navLinks} filterLinks={filterLinks} />

          {children}

          <footer className={styles.footer}>
            <Link href="/legal/terms">{dictionary.footer.terms}</Link>
            <Link href="/legal/privacy">{dictionary.footer.privacy}</Link>
            <Link href="/legal/impressum">{dictionary.footer.imprint}</Link>
            <Link href="/legal/contact">{dictionary.footer.contact}</Link>
            <LanguageSwitcher />
          </footer>
        </I18nProvider>
      </body>
    </html>
  );
}