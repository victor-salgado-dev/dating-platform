import type { Metadata } from 'next';
import { cookies } from 'next/headers';
import { Fraunces, Public_Sans } from 'next/font/google';
import Link from 'next/link';
import HeaderChrome, { type NavLinkItem, type FilterLinkItem } from './header-chrome';
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

const NAV_LINKS: NavLinkItem[] = [
  { href: '/', label: 'Inicio' },
  { href: '/discover', label: 'Descubrir' },
  { href: '/activity', label: 'Actividad' },
  { href: '/likes', label: 'Likes' },
  { href: '/visits', label: 'Visitas' },
  { href: '/matches', label: 'Matches' },
  { href: '/messages', label: 'Mensajes' },
];

const FILTER_LINKS: FilterLinkItem[] = [
  { href: '/', label: 'Populares', active: true },
  { href: '/online', label: 'En línea', online: true },
  { href: '/new', label: 'Nuevos miembros' },
  { href: '/favorites', label: 'Favoritos' },
  { href: '/search', label: 'Búsqueda avanzada' },
];

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = cookies();
  const cookieLocale = cookieStore.get(LOCALE_COOKIE_NAME)?.value ?? '';
  const locale: Locale = isValidLocale(cookieLocale) ? cookieLocale : DEFAULT_LOCALE;
  const dictionary = getDictionary(locale);

  return (
    <html lang={locale} className={`${fraunces.variable} ${publicSans.variable}`}>
      <body>
        <I18nProvider locale={locale} dictionary={dictionary}>
          <HeaderChrome navLinks={NAV_LINKS} filterLinks={FILTER_LINKS} />

          {children}

          <footer className={styles.footer}>
            <Link href="/legal/terms">Términos</Link>
            <Link href="/legal/privacy">Privacidad</Link>
            <Link href="/legal/impressum">Aviso legal</Link>
            <Link href="/legal/contact">Contacto</Link>
          </footer>
        </I18nProvider>
      </body>
    </html>
  );
}
