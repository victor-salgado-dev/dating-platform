import type { Metadata } from 'next';
import { Fraunces, Public_Sans } from 'next/font/google';
import Link from 'next/link';
import HeaderChrome, { type NavLinkItem, type FilterLinkItem } from './header-chrome';
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
  return (
    <html lang="es" className={`${fraunces.variable} ${publicSans.variable}`}>
      <body>
        <HeaderChrome navLinks={NAV_LINKS} filterLinks={FILTER_LINKS} />

        {children}

        <footer className={styles.footer}>
          <Link href="/legal/terms">Términos</Link>
          <Link href="/legal/privacy">Privacidad</Link>
          <Link href="/legal/impressum">Aviso legal</Link>
          <Link href="/legal/contact">Contacto</Link>
        </footer>
      </body>
    </html>
  );
}
