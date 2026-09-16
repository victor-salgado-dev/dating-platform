import type { Metadata } from 'next';
import Link from 'next/link';
import AccountNav from './account-nav';
import './globals.css';
import styles from './layout.module.css';

export const metadata: Metadata = {
  title: 'Dating Platform',
  description: 'Plataforma internacional de dating/relaciones',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="es">
      <body>
        {/* Barra superior (Roja) */}
        <header className={styles.headerTop}>
          <Link href="/" className={styles.brand}>
            🤍 Dating Platform
          </Link>

          <nav className={styles.topNav}>
            <Link href="/">Inicio</Link>
            <Link href="/messages">Mensajes</Link>
            <Link href="/matches">Matches</Link>
            <Link href="/likes">Likes</Link>
            <Link href="/activity">Actividad</Link>
          </nav>

          <div className={styles.account}>
            <AccountNav />
          </div>
        </header>

        {/* Barra secundaria de filtros (Blanca) */}
        <nav className={styles.headerBottom}>
          <Link href="/discover">Descubrir</Link>
          <Link href="/favorites">Favoritos</Link>
          <Link href="/" className={styles.activeTab}>⭐ Populares</Link>
          <Link href="/online">En línea</Link>
          <Link href="/new">Nuevos miembros</Link>
        </nav>

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